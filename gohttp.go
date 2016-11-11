package gohttp

import (
	"net/http"
	"path/filepath"
)

// Client is the main struct that wraps net/http
type Client struct {
	c *http.Client

	// request parameters
	query   map[string]string
	headers map[string]string
	url     string
	path    string
}

// DefaultClient provides a simple usable client, it is given for quick usage.
// For more control, please create a client manually.
var DefaultClient = New()

// New returns a new GoClient
func New() *Client {
	return &Client{
		c:       &http.Client{},
		query:   make(map[string]string),
		headers: make(map[string]string),
	}
}

func (c *Client) prepareRequest(method string) (*http.Request, error) {
	req, err := http.NewRequest(method, c.url, nil)

	// concatenate path to url if exists
	if c.path != "" {
		p := req.URL.Path
		req.URL.Path = filepath.Join(p, c.path)
	}

	// setup the query string
	if len(c.query) != 0 {
		q := req.URL.Query()
		for key, value := range c.query {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	// setup headers
	if len(c.headers) != 0 {
		for key, value := range c.headers {
			req.Header.Add(key, value)
		}
	}

	if err != nil {
		return nil, err
	}

	return req, nil
}

// Get handles HTTP GET request, and return response to user
func (c *Client) Get(url string) (*http.Response, error) {
	c.url = url
	req, err := c.prepareRequest("GET")
	if err != nil {
		return nil, err
	}
	return c.c.Do(req)
}

// Path concatenate url with resource path string
func (c *Client) Path(path string) *Client {
	if path != "" {
		c.path = filepath.Join(c.path, path)
	}
	return c
}

// Query set parameter query string
func (c *Client) Query(key, value string) *Client {
	c.query[key] = value
	return c
}

// Header sets request header data
func (c *Client) Header(key, value string) *Client {
	c.headers[key] = value
	return c
}

// Get provides a shortcut to use
func Get(url string) (*http.Response, error) {
	return DefaultClient.Get(url)
}
