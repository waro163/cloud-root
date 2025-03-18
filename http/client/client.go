package client

import (
	"io"
	"net/http"
	"net/url"
	"time"

	cloudroot "github.com/waro163/cloud-root"
	"github.com/waro163/requests"
)

type BaseClient struct {
	*requests.Client
}

func NewClient(cfg cloudroot.OutgoingService, opts ...Options) (*BaseClient, error) {
	baseUrl, err := url.Parse(cfg.Host)
	if err != nil {
		return nil, err
	}

	cli := requests.NewClient(
		requests.WithName(cfg.Name),
		requests.WithBaseURL(baseUrl),
		requests.WithTimeout(time.Duration(cfg.Timeout)*time.Second),
	)
	if cfg.DefaultRequestHeaders != nil {
		defaultHeaders := http.Header{}
		for _, header := range cfg.DefaultRequestHeaders {
			defaultHeaders.Add(header.Key, header.Value)
		}
		cli.WithDefaultRequestHeaders(defaultHeaders)
	}
	baseClient := &BaseClient{
		Client: cli,
	}
	for _, opt := range opts {
		opt(baseClient)
	}
	return baseClient, nil
}

func (cli *BaseClient) DoBaseRequest(req *http.Request) (*http.Response, []byte, error) {
	var resp *http.Response

	resp, err := cli.Do(req)
	if err != nil {
		return nil, nil, err
	}
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return resp, nil, err
	}
	return resp, body, nil
}

func (cli *BaseClient) DoBaseRequestRetry(req *http.Request, n int) (*http.Response, []byte, error) {
	if n <= 0 {
		n = 3
	}
	var resp *http.Response
	var body []byte
	var err error
	for i := 0; i < n; i++ {
		resp, body, err = cli.DoBaseRequest(req)
		if err != nil {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			continue
		}
		return resp, body, nil
	}
	return resp, body, err
}
