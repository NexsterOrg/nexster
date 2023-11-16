package blob

import (
	"context"
	"fmt"

	blb "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"

	azblob "github.com/NamalSanjaya/nexster/pkgs/azure/blob_storage"
	"github.com/NamalSanjaya/nexster/pkgs/utill/uuid"
)

type blobClient struct {
	azContainerName string
	azClient        azblob.Interface
}

var _ Interface = (*blobClient)(nil)

func NewBlobClient(container string, azBlobIntfce azblob.Interface) *blobClient {
	return &blobClient{
		azContainerName: container,
		azClient:        azBlobIntfce,
	}
}

func (bc *blobClient) ImageReader(ctx context.Context, blobName string) (*blb.RetryReader, string, error) {
	return bc.azClient.DownloadBlob(ctx, bc.azContainerName, blobName)
}

// typeName: png, jpg, jpeg etc
func (bc *blobClient) UploadImage(ctx context.Context, typeName string, data []byte, options *UploadImageOptions) (string, error) {
	var imgFullName string
	if options.BlobName == "" {
		imgFullName = mkImgFullname(uuid.GenUUID4(), typeName)
	} else {
		imgFullName = mkImgFullname(options.BlobName, typeName)
	}
	err := bc.azClient.UploadBuffer(ctx, bc.azContainerName, imgFullName, mkImgContentType(typeName), data)
	if err != nil {
		return "", err
	}
	return imgFullName, nil
}

func mkImgFullname(name, typeName string) string {
	return fmt.Sprintf("%s.%s", name, typeName)
}

func mkImgContentType(typeName string) string {
	return fmt.Sprintf("image/%s", typeName)
}
