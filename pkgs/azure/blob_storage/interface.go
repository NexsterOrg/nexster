package azure

import (
	"context"

	blb "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
)

type Interface interface {
	DownloadBlob(ctx context.Context, container, blob string) (*blb.RetryReader, string, error)
}
