package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
)

type AzBlobClientConfigs struct {
	StorageAccount string            `yaml:"storageAccount"`
	Containers     map[string]string `yaml:"containers"`
}

type AzBlobClient struct {
	client *azblob.Client
}

var _ Interface = (*AzBlobClient)(nil)

func NewAzBlobClient(cfgs *AzBlobClientConfigs) *AzBlobClient {
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		panic(fmt.Errorf("failed to start azure client: failed to create DefaultAzureCredential: %v", err))
	}
	client, err := azblob.NewClient(fmt.Sprintf("https://%s.blob.core.windows.net/", cfgs.StorageAccount), credential, nil)
	if err != nil {
		panic(fmt.Errorf("failed to start azure client: %v", err))
	}
	return &AzBlobClient{client: client}
}

// close after utilizing io.Reader
func (azbc *AzBlobClient) DownloadBlob(ctx context.Context, container, blob string) (*blob.RetryReader, string, error) {
	get, err := azbc.client.DownloadStream(ctx, container, blob, nil)
	if err != nil {
		return nil, "", err
	}
	return get.NewRetryReader(ctx, &azblob.RetryReaderOptions{}), *get.ContentType, nil
}

func (azbc *AzBlobClient) UploadBuffer(ctx context.Context, containerName, blobName, contentType string, data []byte) error {

	_, err := azbc.client.UploadBuffer(ctx, containerName, blobName, data,
		&blockblob.UploadBufferOptions{HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: to.Ptr(contentType),
		}})

	return err
}
