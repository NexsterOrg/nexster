package blob

import (
	"context"

	blb "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
)

type Interface interface {
	ImageReader(ctx context.Context, blobName string) (*blb.RetryReader, string, error)
	UploadImage(ctx context.Context, typeName string, data []byte, options *UploadImageOptions) (string, error)
}

type UploadImageOptions struct {
	BlobName string
}
