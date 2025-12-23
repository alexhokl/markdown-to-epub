package cmd

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexhokl/helper/cli"
	"github.com/alexhokl/helper/iohelper"
	"github.com/go-shiori/go-epub"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

//go:embed style.css
var defaultCSS string

type generateOptions struct {
	markdownFilename string
	epubFilename     string
	overwrite        bool
	title            string
	author           string
	language         string
}

var generateOps generateOptions

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate epub file from the specified markdown file",
	RunE:  runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	flags := generateCmd.Flags()
	flags.StringVarP(&generateOps.markdownFilename, "input", "i", "", "Path to markdown file")
	flags.StringVarP(&generateOps.epubFilename, "output", "o", "", "Path to output epub file")
	flags.BoolVarP(&generateOps.overwrite, "overwrite", "f", false, "Overwrite existing epub file")
	flags.StringVarP(&generateOps.title, "title", "t", "", "Title of the book (defaults to filename)")
	flags.StringVarP(&generateOps.author, "author", "a", "", "Author of the book")
	flags.StringVarP(&generateOps.language, "language", "l", "en", "Language code (e.g., en, ja, zh)")

	if err := generateCmd.MarkFlagRequired("input"); err != nil {
		cli.LogUnableToMarkFlagAsRequired("input", err)
	}
	if err := generateCmd.MarkFlagRequired("output"); err != nil {
		cli.LogUnableToMarkFlagAsRequired("output", err)
	}
}

func runGenerate(cmd *cobra.Command, args []string) error {
	if err := validateGenerateOptions(generateOps); err != nil {
		return err
	}

	// Read the Markdown file
	content, err := os.ReadFile(generateOps.markdownFilename)
	if err != nil {
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	// Convert Markdown to HTML
	htmlContent, err := convertMarkdownToHTML(content)
	if err != nil {
		return fmt.Errorf("failed to convert markdown to HTML: %w", err)
	}

	// Determine title
	title := generateOps.title
	if title == "" {
		// Try to extract title from first H1 heading
		title = extractTitleFromMarkdown(string(content))
		if title == "" {
			// Fall back to filename without extension
			title = strings.TrimSuffix(filepath.Base(generateOps.markdownFilename), filepath.Ext(generateOps.markdownFilename))
		}
	}

	// Create ePub
	if err := createEpub(title, htmlContent); err != nil {
		return fmt.Errorf("failed to create epub: %w", err)
	}

	fmt.Printf("Successfully created %s\n", generateOps.epubFilename)
	return nil
}

func validateGenerateOptions(options generateOptions) error {
	if !iohelper.IsFileExist(options.markdownFilename) {
		return fmt.Errorf("markdown file %s does not exist", options.markdownFilename)
	}

	if iohelper.IsFileExist(options.epubFilename) && !options.overwrite {
		return fmt.Errorf("epub file %s already exists, use option -f to overwrite", options.epubFilename)
	}

	return nil
}

func convertMarkdownToHTML(content []byte) (string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert(content, &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func extractTitleFromMarkdown(content string) string {
	lines := strings.SplitSeq(content, "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "# "); ok {
			return after
		}
	}
	return ""
}

func createEpub(title, htmlContent string) error {
	// Create a new ePub
	e, err := epub.NewEpub(title)
	if err != nil {
		return fmt.Errorf("failed to create epub: %w", err)
	}

	// Set metadata
	e.SetLang(generateOps.language)
	if generateOps.author != "" {
		e.SetAuthor(generateOps.author)
	}

	var cssPath string

	// Use embedded CSS
	css := defaultCSS

	// Write CSS to a temporary file (go-epub requires a file path or URL)
	tmpFile, err := os.CreateTemp("", "epub-style-*.css")
	if err != nil {
		return fmt.Errorf("failed to create temp CSS file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(css); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write CSS to temp file: %w", err)
	}
	tmpFile.Close()

	// Add CSS to ePub
	cssPath, err = e.AddCSS(tmpFile.Name(), "style.css")
	if err != nil {
		return fmt.Errorf("failed to add CSS: %w", err)
	}

	// Add cover page as the first section
	coverHTML := generateCoverPage(title)
	_, err = e.AddSection(coverHTML, "Cover", "cover.xhtml", cssPath)
	if err != nil {
		return fmt.Errorf("failed to add cover page: %w", err)
	}

	// Add the content as a section
	_, err = e.AddSection(htmlContent, title, "", cssPath)
	if err != nil {
		return fmt.Errorf("failed to add section: %w", err)
	}

	// Write the ePub file
	if err := e.Write(generateOps.epubFilename); err != nil {
		return fmt.Errorf("failed to write epub file: %w", err)
	}

	return nil
}

// generateCoverPage creates an HTML cover page with the book title
func generateCoverPage(title string) string {
	return fmt.Sprintf(`<div class="cover-page">
	<h1 class="cover-title">%s</h1>
</div>`, title)
}
