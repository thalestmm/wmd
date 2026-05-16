package main

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
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/websocket"
	"github.com/thalestmm/wmd/templates"
)

var (
	filePath string
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <path-to-markdown-file>", os.Args[0])
	}
	filePath = os.Args[1]

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf("File does not exist: %s", filePath)
	}

	listener, err := availableListener(8080)
	if err != nil {
		log.Fatalf("No available port: %v", err)
	}

	go watchFile(filePath)

	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/ws", serveWS)

	addr := listener.Addr().String()
	url := "http://" + addr
	fmt.Printf("Watching: %s\n", filePath)
	fmt.Printf("Listening on %s\n", url)

	go openBrowser(url)

	log.Fatal(http.Serve(listener, nil))
}

// availableListener tries startPort and increments until a free port is found.
func availableListener(startPort int) (net.Listener, error) {
	for port := startPort; port < startPort+100; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		if err == nil {
			return ln, nil
		}
	}
	return nil, fmt.Errorf("no free port in range %d-%d", startPort, startPort+99)
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	mdContent, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	extensions := parser.CommonExtensions | parser.MathJax
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(mdContent)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	renderedHTML := markdown.Render(doc, renderer)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.Page(template.HTML(renderedHTML)).Render(r.Context(), w)
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

func openBrowser(url string) {
	// Wait briefly for the server to start before opening.
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
