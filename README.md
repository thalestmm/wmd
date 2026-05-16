# wmd

Live markdown preview with hot reload and PDF export. Renders with LaTeX typography and MathJax support.

## Installation

```bash
go install github.com/thalestmm/wmd@latest
```

Requires [Chrome or Chromium](https://www.chromium.org/getting-started/downloading-chromium/) for PDF export.

## Usage

### Live preview

```bash
wmd example.md
# or explicitly:
wmd serve example.md
```

Opens `http://localhost:8080` (or the next available port) in your browser. The page hot-reloads on every file save. Works with editors that do atomic saves (Neovim, Vim, etc.).

### Export to PDF

```bash
# saves example.pdf next to the source file
wmd pdf example.md

# custom output path
wmd pdf example.md -o report.pdf
```

PDF output matches what you see in the browser: A4 page, LaTeX typography, rendered math.

You can also export from the browser while the preview server is running — click the **Download PDF** button in the top-right corner.

## Markdown features

- Standard CommonMark
- LaTeX math via MathJax: inline `$...$` and display `$$...$$`
- Tables, strikethrough, autolinks (CommonMark extensions)
- External links open in a new tab
