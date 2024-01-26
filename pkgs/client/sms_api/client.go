package smsapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	uh "github.com/NamalSanjaya/nexster/pkgs/utill/http"
)

type SmsClientConfig struct {
	Host      string `yaml:"host"`
	Path      string `yaml:"path"`
	ApiKey    string `yaml:"apiKey"`
	ApiSecret string `yaml:"apiSecret"`
}

type smsApiClient struct {
	client    *http.Client
	domain    string // https://rest.nexmo.com
	path      string // /sms/json
	apiKey    string
	apiSecret string
}

func NewApiClient(config *SmsClientConfig) *smsApiClient {
	return &smsApiClient{
		client:    &http.Client{},
		domain:    fmt.Sprintf("https://%s", config.Host),
		path:      config.Path,
		apiKey:    config.ApiKey,
		apiSecret: config.ApiSecret,
	}
}

func (sac *smsApiClient) SendSms(ctx context.Context, from, msg, to string) (err error) {
	formData := url.Values{
		"from":       {from},
		"text":       {msg},
		"to":         {to},
		"api_key":    {sac.apiKey},
		"api_secret": {sac.apiSecret},
	}
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s%s", sac.domain, sac.path), strings.NewReader(formData.Encode()))
	if err != nil {
		err = fmt.Errorf("failed to prepare http req: %v", err)
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := sac.client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to send the http req: %v", err)
		return
	}
	defer resp.Body.Close()

	if uh.IsNon2xxStatusCode(resp.StatusCode) {
		err = fmt.Errorf("request failed with %s code", resp.Status)
		return
	}
	return
}
