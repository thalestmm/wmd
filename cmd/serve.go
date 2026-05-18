package cmd

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	ipdf "github.com/thalestmm/wmd/internal/pdf"
	"github.com/thalestmm/wmd/internal/render"
	"github.com/thalestmm/wmd/templates"
)

var (
	serveFilePath string
	serverAddr    string
	autoFormat    bool
	upgrader      = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	clients       = make(map[*websocket.Conn]bool)
	clientsMu     sync.Mutex
)

var serveCmd = &cobra.Command{
	Use:     "serve <file.md>",
	Short:   "Start hot-reload preview server",
	Aliases: []string{"s", "watch", "w"},
	Args:    cobra.ExactArgs(1),
	RunE:    runServe,
}

func init() {
	serveCmd.Flags().BoolVarP(&autoFormat, "auto-format", "f", false, "format file on save")
}

func runServe(_ *cobra.Command, args []string) error {
	serveFilePath = args[0]
	if _, err := os.Stat(serveFilePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", serveFilePath)
	}

	ln, err := availableListener(8080)
	if err != nil {
		return fmt.Errorf("no available port: %w", err)
	}
	serverAddr = ln.Addr().String()

	go watchFile(serveFilePath)

	mux := http.NewServeMux()
	mux.HandleFunc("/", serveHTML)
	mux.HandleFunc("/ws", serveWS)
	mux.HandleFunc("/pdf", handlePDF)

	url := "http://" + serverAddr
	fmt.Printf("Watching: %s\n", serveFilePath)
	fmt.Printf("Listening on %s\n", url)

	go openBrowser(url)

	return http.Serve(ln, mux)
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	mdContent, err := os.ReadFile(serveFilePath)
	if err != nil {
		http.Error(w, "error reading file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.Page(template.HTML(render.Markdown(mdContent))).Render(r.Context(), w)
}

func handlePDF(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Generating PDF...")
	pdfBytes, err := ipdf.FromURL("http://" + serverAddr + "/")
	if err != nil {
		http.Error(w, "PDF generation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	base := filepath.Base(serveFilePath)
	outName := base[:len(base)-len(filepath.Ext(base))] + ".pdf"
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="`+outName+`"`)
	w.Write(pdfBytes)
	fmt.Println("PDF sent.")
}

func serveWS(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer ws.Close()
	clientsMu.Lock()
	clients[ws] = true
	clientsMu.Unlock()
	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			break
		}
	}
	clientsMu.Lock()
	delete(clients, ws)
	clientsMu.Unlock()
}

func watchFile(target string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	// Watch dir — editors like Neovim do atomic saves (rename/create), breaking direct file watches.
	dir := filepath.Dir(target)
	if err = watcher.Add(dir); err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if filepath.Clean(event.Name) == filepath.Clean(target) {
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					if autoFormat {
						if err := formatFile(target); err != nil {
							log.Println("Auto-format error:", err)
						}
					}
					broadcastReload()
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
	}
}

func broadcastReload() {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, []byte("reload")); err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}

func availableListener(startPort int) (net.Listener, error) {
	for port := startPort; port < startPort+100; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		if err == nil {
			return ln, nil
		}
	}
	return nil, fmt.Errorf("no free port in range %d–%d", startPort, startPort+99)
}

func openBrowser(url string) {
	time.Sleep(300 * time.Millisecond)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return
	}
	cmd.Start()
}
