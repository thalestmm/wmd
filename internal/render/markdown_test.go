package render

import (
	"bytes"
	"testing"
)

func TestMermaidFenceRendersFigure(t *testing.T) {
	out := Markdown([]byte("```mermaid\nflowchart LR\nA --> B\n```\n"))
	if !bytes.Contains(out, []byte(`<figure class="mermaid-figure">`)) {
		t.Fatalf("expected mermaid figure, got:\n%s", out)
	}
	if !bytes.Contains(out, []byte(`<pre class="mermaid">`)) {
		t.Fatalf("expected mermaid source container, got:\n%s", out)
	}
	if !bytes.Contains(out, []byte("flowchart LR")) {
		t.Fatalf("expected diagram source, got:\n%s", out)
	}
	if bytes.Contains(out, []byte("language-mermaid")) {
		t.Fatalf("mermaid should not render as a highlighted code block:\n%s", out)
	}
}

func TestMermaidArrowsSurviveHTMLEscape(t *testing.T) {
	out := Markdown([]byte("```mermaid\nflowchart LR\nA --> B\n```\n"))
	if !bytes.Contains(out, []byte("A --&gt; B")) {
		t.Fatalf("arrows should be HTML-escaped in the DOM, got:\n%s", out)
	}
}

func TestPlainCodeFenceUnchanged(t *testing.T) {
	out := Markdown([]byte("```go\nfmt.Println()\n```\n"))
	if bytes.Contains(out, []byte("mermaid")) {
		t.Fatalf("go fence should not be treated as mermaid:\n%s", out)
	}
	if !bytes.Contains(out, []byte("<pre>")) {
		t.Fatalf("expected pre/code, got:\n%s", out)
	}
}
