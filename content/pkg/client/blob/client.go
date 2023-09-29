package blob

import (
	"context"

	blb "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"

	azblob "github.com/NamalSanjaya/nexster/pkgs/azure/blob_storage"
)

type blobClient struct {
	azContainerName string
	azClient        azblob.Interface
}

func NewBlobClient(container string, azBlobIntfce azblob.Interface) *blobClient {
	return &blobClient{
		azContainerName: container,
		azClient:        azBlobIntfce,
	}
}

func (bc *blobClient) ImageReader(ctx context.Context, blobName string) (*blb.RetryReader, string, error) {
	return bc.azClient.DownloadBlob(ctx, bc.azContainerName, blobName)
}
