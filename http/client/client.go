package client

import (
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
