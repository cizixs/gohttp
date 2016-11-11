package gohttp

import (
	"net/http"
)

// Client is the main struct that wraps net/http
type Client struct {
	req   *http.Request
	c     *http.Client
	query map[string]string
	url   string
}

// DefaultClient provides a simple usable client, it is given for quick usage.
// For more control, please create a client manually.
var DefaultClient = New()

// New returns a new GoClient
func New() *Client {
	return &Client{
		c:     &http.Client{},
		query: make(map[string]string),
	}
}

func (g *Client) prepareRequest(method string) (*http.Request, error) {
	req, err := http.NewRequest(method, g.url, nil)

	if len(g.query) != 0 {
		q := req.URL.Query()
		for key, value := range g.query {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	if err != nil {
		return nil, err
	}

	return req, nil
}

// Get handles HTTP GET request, and return response to user
func (g *Client) Get(url string) (*http.Response, error) {
	g.url = url
	req, err := g.prepareRequest("GET")
	if err != nil {
		return nil, err
	}
	return g.c.Do(req)
}

// Query set parameter query string
func (g *Client) Query(key, value string) *Client {
	g.query[key] = value
	return g
}

// Get provides a shortcut to use
func Get(url string) (*http.Response, error) {
	return DefaultClient.Get(url)
}
