# markdown-to-epub

A command-line tool to convert Markdown files to EPUB format.

## Usage

```bash
markdown-to-epub generate -i input.md -o output.epub
```

### Options

- `-i, --input` - Path to the markdown file (required)
- `-o, --output` - Path to output epub file (required)
- `-t, --title` - Title of the book (defaults to first H1 heading or filename)
- `-l, --language` - Language code, e.g., `en`, `ja`, `zh` (default: `en`)
- `-f, --overwrite` - Overwrite existing epub file

## Japanese Language Support

This tool includes the embedded Noto Sans JP font for proper Japanese character
rendering. When generating Japanese content, use the `-l ja` flag:

```bash
markdown-to-epub generate -i japanese.md -o output.epub -l ja
```
