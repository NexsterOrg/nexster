package contentapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	cl "github.com/NamalSanjaya/nexster/pkgs/client"
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}
	return data["url"], nil
}
