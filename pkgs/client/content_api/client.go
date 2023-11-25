package contentapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	cl "github.com/NamalSanjaya/nexster/pkgs/client"
	uh "github.com/NamalSanjaya/nexster/pkgs/utill/http"
)

type conentApiClient struct {
	client *http.Client
	domain string // http://127.0.0.1:8002
}

func NewApiClient(config *cl.HttpClientConfig) *conentApiClient {
	return &conentApiClient{
		client: &http.Client{},
		domain: fmt.Sprintf("http://%s:%d", config.Host, config.Port),
	}
}

func (ca *conentApiClient) CreateImageUrl(imgIdWithNamespace, permission string) (string, error) {
	var data map[string]string
	resp, err := ca.client.Get(fmt.Sprintf("%s/content/hmac/image/%s?perm=%s", ca.domain, imgIdWithNamespace, permission))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if uh.IsNon2xxStatusCode(resp.StatusCode) {
		return "", fmt.Errorf("request failed with %s code", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}
	return data["url"], nil
}

func (ca *conentApiClient) DeleteImage(ctx context.Context, imgIdWithNamespace string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/content/images/%s", ca.domain, imgIdWithNamespace), nil)
	if err != nil {
		return err
	}

	resp, err := ca.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if uh.IsNon2xxStatusCode(resp.StatusCode) {
		return fmt.Errorf("request failed with %s code", resp.Status)
	}
	return nil
}

// helper functions

func (ca *conentApiClient) GetPermission(ownerKey, viewerKey string) string {
	if ownerKey == viewerKey {
		return Owner
	}
	return Viewer
}

func GetMediaKey(imgFullname string) string {
	parts := strings.Split(imgFullname, ".")
	if len(parts) != 2 {
		return ""
	}
	parts2 := strings.Split(parts[0], "/")
	if len(parts2) != 2 {
		return ""
	}
	return parts2[1]
}
