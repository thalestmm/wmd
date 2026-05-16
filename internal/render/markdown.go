package render

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func Markdown(src []byte) []byte {
	extensions := parser.CommonExtensions | parser.MathJax
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(src)
	opts := html.RendererOptions{Flags: html.CommonFlags | html.HrefTargetBlank}
	return markdown.Render(doc, html.NewRenderer(opts))
}
