package loader

import (
	"context"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/schema"
)

type Loader struct {
}

func NewLoader() (document.Loader, error) {
	return nil, nil
}

func Load(ctx context.Context, src document.Source, opts ...document.LoaderOption) ([]*schema.Document, error) {
	//Todo:
	return nil, nil
}

