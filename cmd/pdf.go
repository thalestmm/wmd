package cmd

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	ipdf "github.com/thalestmm/wmd/internal/pdf"
	"github.com/thalestmm/wmd/internal/render"
	"github.com/thalestmm/wmd/templates"
)

var pdfOutput string

var pdfCmd = &cobra.Command{
	Use:     "pdf <file.md>",
	Short:   "Render markdown to PDF",
	Aliases: []string{"render", "export"},
	Args:    cobra.ExactArgs(1),
	RunE:    runPDF,
}

func init() {
	pdfCmd.Flags().StringVarP(&pdfOutput, "output", "o", "", "output PDF path (default: <input>.pdf)")
}

func runPDF(_ *cobra.Command, args []string) error {
	inputFile := args[0]
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", inputFile)
	}

	outputFile := pdfOutput
	if outputFile == "" {
		base := filepath.Base(inputFile)
		outputFile = base[:len(base)-len(filepath.Ext(base))] + ".pdf"
	}

	// Spin up a temp server so chromedp can load external CSS/JS from CDN.
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return fmt.Errorf("failed to bind: %w", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mdContent, err := os.ReadFile(inputFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		templates.Page(template.HTML(render.Markdown(mdContent))).Render(r.Context(), w)
	})
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	defer srv.Close()

	fmt.Printf("Generating PDF: %s → %s\n", inputFile, outputFile)

	pdfBytes, err := ipdf.FromURL("http://" + ln.Addr().String() + "/")
	if err != nil {
		return fmt.Errorf("PDF generation failed: %w", err)
	}

	if err := os.WriteFile(outputFile, pdfBytes, 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	fmt.Printf("Saved: %s\n", outputFile)
	return nil
}
