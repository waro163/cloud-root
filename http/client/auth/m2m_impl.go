package auth

import (
	"context"
	"encoding/json"
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

type M2MRequestBody struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Resource     string `json:"resource"`
	Scope        string `json:"scope"`
}
type M2MResponseBody struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
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
		return nil
	}

	m.mux.Lock()
	defer m.mux.Unlock()

	token, err = m.cache.GetString(m.cacheKey)
	if err == nil && token != "" {
		req.Header.Add(authHeader, prefix+token)
	}
	//
	resp, err := m.client.Post(req.Context(), "", &requests.HTTPOptions{
		Body: M2MRequestBody{
			ClientId:     m.clientId,
			ClientSecret: m.clientSecret,
			Resource:     m.resource,
			Scope:        m.scope,
		},
	})
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("empty response when get m2m token")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}
	var res M2MResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}
	duration := time.Duration(res.ExpiresIn-300) * time.Second
	m.cache.Set(m.cacheKey, res.AccessToken, duration)
	req.Header.Add(authHeader, prefix+token)
	return nil
}

func (m *m2mAuthSet) Reset(ctx context.Context) {
	m.cache.Delete(m.cacheKey)
	return
}
