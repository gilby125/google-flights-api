package hotels

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/browserutils/kooky"
	"github.com/hashicorp/go-retryablehttp"
)

// Map is safe for concurrent use by multiple goroutines.
type Map[K comparable, V any] struct {
	m sync.Map
}

func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	v, ok := m.m.Load(key)
	if !ok {
		return value, ok
	}
	return v.(V), ok
}

func (m *Map[K, V]) Store(key K, value V) { m.m.Store(key, value) }

type userAgentTransport struct {
	base http.RoundTripper
	ua   string
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", t.ua)
	return t.base.RoundTrip(req)
}

type httpClient interface {
	Do(req *retryablehttp.Request) (*http.Response, error)
}

type Session struct {
	// Cache for location resolution if needed
	Locations Map[string, string]

	client  httpClient
	cookies []string
}

func customRetryPolicy() func(ctx context.Context, resp *http.Response, err error) (bool, error) {
	return func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
				return false, ctx.Err()
			}
		}

		if resp == nil {
			return true, fmt.Errorf("response is nil")
		}

		if resp.StatusCode != http.StatusOK {
			return true, fmt.Errorf("wrong status code: %d", resp.StatusCode)
		}
		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}
}

func getCookies(res *http.Response) ([]string, error) {
	var cookies []string
	if setCookie, ok := res.Header["Set-Cookie"]; ok {
		for _, c := range setCookie {
			cookies = append(cookies, strings.Split(c, ";")[0])
		}
		return cookies, nil
	}
	return nil, fmt.Errorf("could not find the 'Set-Cookie' header in the initialization response")
}

func New() (*Session, error) {
	client := retryablehttp.NewClient()
	client.RetryMax = 5
	client.Logger = nil
	client.CheckRetry = customRetryPolicy()
	client.RetryWaitMin = time.Second
	client.HTTPClient.Timeout = 90 * time.Second

	// Set a modern User-Agent
	client.HTTPClient.Transport = &userAgentTransport{
		base: http.DefaultTransport,
		ua:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}

	// Initialize cookies from google.com
	res, err := client.Get("https://www.google.com/")
	if err != nil {
		return nil, fmt.Errorf("new session: err sending request to www.google.com: %v", err)
	}
	defer res.Body.Close()

	cookies, err := getCookies(res)
	if err != nil {
		return nil, fmt.Errorf("new session: err getting cookies: %v", err)
	}

	GOOGLE_ABUSE_EXEMPTION := kooky.ReadCookies(kooky.Valid, kooky.DomainHasSuffix(`google.com`), kooky.Name(`GOOGLE_ABUSE_EXEMPTION`))

	if len(GOOGLE_ABUSE_EXEMPTION) == 1 {
		cookies = append(cookies, GOOGLE_ABUSE_EXEMPTION[0].Value)
	}

	return &Session{
		Locations: Map[string, string]{},
		client:    client,
		cookies:   cookies,
	}, nil
}
