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

// type blobReader struct{}

// func newBlobReader() *blobReader {
// 	return &blobReader{}
// }

// func (br *blobReader) Read(p []byte) (n int, err error){
// 	return br.Read(p)
// }

// func (br *blobReader) Close() error {
// 	return br.Close()
// }

func NewBlobClient(azContainerName string, azBlobIntfce azblob.Interface) *blobClient {
	return &blobClient{
		azContainerName: azContainerName,
		azClient:        azBlobIntfce,
	}
}

func (bc *blobClient) ImageReader(ctx context.Context, blobName string) (*blb.RetryReader, string, error) {
	return bc.azClient.DownloadBlob(ctx, bc.azContainerName, blobName)
}
