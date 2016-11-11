package gohttp

import (
	"net/http"
	"path/filepath"

	"github.com/google/go-querystring/query"
)

// Client is the main struct that wraps net/http
type Client struct {
	c *http.Client

	// request parameters
	query        map[string]string
	queryStructs []interface{}
	headers      map[string]string
	url          string
	path         string
}

// DefaultClient provides a simple usable client, it is given for quick usage.
// For more control, please create a client manually.
var DefaultClient = New()

// New returns a new GoClient
func New() *Client {
	return &Client{
		c:            &http.Client{},
		query:        make(map[string]string),
		queryStructs: make([]interface{}, 5),
		headers:      make(map[string]string),
	}
}

func (c *Client) prepareRequest(method string) (*http.Request, error) {
	req, err := http.NewRequest(method, c.url, nil)

	// concatenate path to url if exists
	if c.path != "" {
		p := req.URL.Path
		req.URL.Path = filepath.Join(p, c.path)
	}

	// setup the query string, there can be many query strings,
	// and they're connected with `&` symbol. Also rawquery and
	// queryStructs data will be merged.
	if len(c.query) != 0 {
		q := req.URL.Query()
		for key, value := range c.query {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	if len(c.queryStructs) != 0 {
		q := req.URL.Query()

		for _, queryStruct := range c.queryStructs {
			qValues, err := query.Values(queryStruct)
			if err != nil {
				return req, err
			}
			for key, values := range qValues {
				for _, value := range values {
					q.Add(key, value)
				}
			}
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

// QueryStruct parses a struct as query strings
// On how it works, please refer to github.com/google/go-querystring repo
func (c *Client) QueryStruct(queryStruct interface{}) *Client {
	c.queryStructs = append(c.queryStructs, queryStruct)
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
