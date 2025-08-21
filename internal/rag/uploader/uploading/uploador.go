package uploading

import "context"

type Uploader interface {
	Upload(ctx context.Context, fileUrl string) ([]string, error)
}
