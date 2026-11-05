package render

import (
	"bytes"
	"io"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func Markdown(src []byte) []byte {
	extensions := parser.CommonExtensions | parser.MathJax
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(src)
	opts := html.RendererOptions{
		Flags:          html.CommonFlags | html.HrefTargetBlank,
		RenderNodeHook: mermaidCodeBlock,
	}
	return markdown.Render(doc, html.NewRenderer(opts))
}

func mermaidCodeBlock(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	code, ok := node.(*ast.CodeBlock)
	if !ok || !isMermaidInfo(code.Info) {
		return ast.GoToNext, false
	}
	if !entering {
		return ast.SkipChildren, true
	}
	io.WriteString(w, `<figure class="mermaid-figure"><pre class="mermaid">`)
	html.EscapeHTML(w, code.Literal)
	io.WriteString(w, `</pre></figure>`)
	if !html.IsListItem(code.Parent) {
		io.WriteString(w, "\n")
	}
	return ast.SkipChildren, true
}

func isMermaidInfo(info []byte) bool {
	fields := bytes.Fields(bytes.ToLower(info))
	return len(fields) > 0 && bytes.Equal(fields[0], []byte("mermaid"))
}
