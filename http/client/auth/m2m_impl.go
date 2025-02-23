package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	cloudroot "github.com/waro163/cloud-root"
	"github.com/waro163/cloud-root/cache"
	"github.com/waro163/requests"
)

const (
	authHeader = "Authorization"
	prefix     = "Bearer "
)

type m2mAuthSet struct {
	client       *requests.Client
	cacheKey     string
	cache        cache.ICache
	mux          *sync.Mutex
	clientId     string
	clientSecret string
	resource     string
	scope        string
}

func NewM2mAuthSet(cfg cloudroot.AuthSetConfig) (IAuthSet, error) {
	baseUrl, err := url.Parse(cfg.Url)
	if err != nil {
		return nil, err
	}

	cli := requests.NewClient(
		requests.WithBaseURL(baseUrl),
		requests.WithTimeout(10*time.Second),
	)
	return &m2mAuthSet{
		client:       cli,
		cacheKey:     fmt.Sprintf("%s_%s_%s", cfg.ClientId, cfg.Resource, cfg.Scope),
		cache:        cache.DefaultMemoryCache,
		mux:          new(sync.Mutex),
		clientId:     cfg.ClientId,
		clientSecret: cfg.ClientSecret,
		resource:     cfg.Resource,
		scope:        cfg.Resource,
	}, nil
}

func (m *m2mAuthSet) SetRequest(req *http.Request) error {
	token, err := m.cache.GetString(m.cacheKey)
	if err == nil && token != "" {
		req.Header.Add(authHeader, prefix+token)
	}
	return nil
}

func (m *m2mAuthSet) Reset(ctx context.Context) {
	m.cache.Delete(m.cacheKey)
	return
}
