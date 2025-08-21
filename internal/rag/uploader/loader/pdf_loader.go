package loader

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudwego/eino-ext/components/document/parser/pdf"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/document/parser"
	"github.com/cloudwego/eino/schema"
	_ "github.com/leebrouse/eino/internal/config"
	"github.com/spf13/viper"
)

type Loader struct {
	toPages bool
}

// NewLoader creates a new PDF Loader.
func NewLoader() (document.Loader, error) {
	return &Loader{
		toPages: viper.GetBool("loader.toPages"),
	}, nil
}

// Load parses a PDF document from the given source.
func (l *Loader) Load(ctx context.Context, src document.Source, opts ...document.LoaderOption) ([]*schema.Document, error) {
	// Initialize PDF parser
	p, err := pdf.NewPDFParser(ctx, &pdf.Config{ToPages: l.toPages})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PDF parser: %w", err)
	}

	// Open PDF file
	file, err := os.Open(src.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF file (%s): %w", src.URI, err)
	}
	defer file.Close()

	// Parse PDF
	docs, err := p.Parse(
		ctx,
		file,
		parser.WithURI(src.URI),
		parser.WithExtraMeta(map[string]any{"source": src.URI}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PDF file (%s): %w", src.URI, err)
	}

	// Log parsing result (optional)
	//fmt.Printf("[Loader] PDF parsed successfully: %d pages\n", len(docs))

	return docs, nil
}
