package generating

import (
	"golang.org/x/net/context"
)

type Generator interface {
	Generate(ctx context.Context, query string) (string, error)
}
