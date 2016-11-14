package gohttp

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-querystring/query"
)

const (
	contentType     = "Content-Type"
	jsonContentType = "application/json"
	formContentType = "application/x-www-form-urlencoded"
)

type basicAuth struct {
	username string
	password string
}

// fileForm wraps one file part of multiple files upload
type fileForm struct {
	fieldName string
	filename  string
	file      *os.File
}

// GoResponse wraps the official `http.Response`, and provides more features.
// The main function is to parse resp body for users.
// In the future, it can gives more information, like request elapsed time,
// redirect history etc
type GoResponse struct {
	*http.Response
}

// AsString returns the response data as string
func (resp *GoResponse) AsString() (string, error) {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// AsBytes return the response body as byte slice
func (resp *GoResponse) AsBytes() ([]byte, error) {
	return ioutil.ReadAll(resp.Body)
}

// AsJSON parses response body to a struct
// Usage:
//    user := &User{}
//    resp.AsJSON(user)
//    fmt.Printf("%s\n", user.Name)
func (resp *GoResponse) AsJSON(v interface{}) error {
	data, err := resp.AsBytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Client is the main struct that wraps net/http
type Client struct {
	c *http.Client

	// query string, query string is a key-value pair follows a
	// url path, and they are encoded in url to escape special characters
	query map[string]string

	// query string from struct data
	queryStructs []interface{}

	// http headers
	headers map[string]string

	// base url, this is sheme + host, but can be any valid url
	url string

	// additional url path, it will be joined with `url` to get the final
	// request url
	path string

	// request body, it should be readble. Body is used in `POST`, `PUT` method to send data
	// to server. How body should be parsed is in the `Content-Type` header.
	// For example, `application/json` means the body is a valid json string,
	// `application/x-www-form-urlencoed` means the body is form data, and it is encoded same as url.
	body io.Reader

	// basic authentication, just plain username and password
	auth basicAuth

	// cookies store request cookie, and send it to server
	cookies []*http.Cookie

	// files represents an array of `os.File` instance, it is used to
	// upload files to server
	files []*fileForm

	// proxy stores proxy url to use. If it is empty, `net/http` will try to load proxy configuration
	// from environment variable.
	proxy string
}

// DefaultClient provides a simple usable client, it is given for quick usage.
// For more control, please create a client manually.
var DefaultClient = New()

// New returns a new GoClient with default values
func New() *Client {
	return &Client{
		query:        make(map[string]string),
		queryStructs: make([]interface{}, 0),
		headers:      make(map[string]string),
		auth:         basicAuth{},
		cookies:      make([]*http.Cookie, 0),
		files:        make([]*fileForm, 0),
	}
}

func (c *Client) prepareRequest(method string) (*http.Request, error) {
	// create the transport and client instance first
	transport := &http.Transport{}
	if c.proxy != "" {
		proxy, err := url.Parse(c.proxy)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxy)
	}
	c.c = &http.Client{Transport: transport}

	// parse files in request.
	// `gohttp` supports posting multiple files, for each file, there should be three component:
	// - fieldName: the upload file button field name in the web. `<input type="file" name="files" multiple>`,
	//   For instance, fieldname would be `files` in this case.
	// - filename: filename string denotes the file to be uploaded. This can be set manually, or extracted from filepath
	// - file content: the actual file data. Ideally, users can pass a filepath, or a os.File instance, or a []byte slice, or a string
	//   as file content
	// On how `multipart/form-data` works, please refer to RFC7578 and RFC 2046.
	if len(c.files) > 0 {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// for each file, write fieldname, filename, and file content to the body
		for _, file := range c.files {
			part, err := writer.CreateFormFile(file.fieldName, file.filename)
			if err != nil {
				return nil, err
			}
			_, err = io.Copy(part, file.file)
			if err != nil {
				return nil, err
			}
		}

		err := writer.Close()
		if err != nil {
			return nil, err
		}

		// finally, get the real content type, and set the header
		multipartContentType := writer.FormDataContentType()
		c.Header(contentType, multipartContentType)
		c.body = body
	}

	// create the basci request
	req, err := http.NewRequest(method, c.url, c.body)

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

	// set cookies
	if len(c.cookies) != 0 {
		for _, cookie := range c.cookies {
			req.AddCookie(cookie)
		}
	}
	// set basic auth if exists
	if c.auth.username != "" && c.auth.password != "" {
		req.SetBasicAuth(c.auth.username, c.auth.password)
	}

	if err != nil {
		return nil, err
	}

	return req, nil
}

// Do takes HTTP and url, then makes the request, return the response
// All other HTTP methods will call `Do` behind the scene, and
// it can be used directly to send the request
// Custom HTTP method can be sent with this method
func (c *Client) Do(method, url string) (*GoResponse, error) {
	c.url = url
	req, err := c.prepareRequest(method)
	if err != nil {
		return nil, err
	}
	resp, err := c.c.Do(req)
	return &GoResponse{resp}, err
}

// Get handles HTTP GET request, and return response to user
// Note that the response is not `http.Response`, but a thin wrapper which does
// exactly what it used to and a little more.
func (c *Client) Get(url string) (*GoResponse, error) {
	return c.Do("GET", url)
}

// Post handles HTTP POST request
func (c *Client) Post(url string) (*GoResponse, error) {
	return c.Do("POST", url)
}

// Head handles HTTP HEAD request
// HEAD request works the same way as GET, except the response body is empty.
func (c *Client) Head(url string) (*GoResponse, error) {
	return c.Do("HEAD", url)
}

// Put handles HTTP PUT request
func (c *Client) Put(url string) (*GoResponse, error) {
	return c.Do("PUT", url)
}

// Delete handles HTTP DELETE request
func (c *Client) Delete(url string) (*GoResponse, error) {
	return c.Do("DELETE", url)
}

// Patch handles HTTP PATCH request
func (c *Client) Patch(url string) (*GoResponse, error) {
	return c.Do("PATCH", url)
}

// Options handles HTTP OPTIONS request
func (c *Client) Options(url string) (*GoResponse, error) {
	return c.Do("OPTIONS", url)
}

// Path concatenates base url with resource path.
// Path can be with or without slash `/` at both end,
// it will be handled properly.
//
// Usage:
//    gohttp.New().Path("users/cizixs").Get("someurl.com")
//
func (c *Client) Path(paths ...string) *Client {
	for _, path := range paths {
		if path != "" {
			c.path = filepath.Join(c.path, path)
		}
	}
	return c
}

// Proxy sets proxy server the client uses.
// If it is empty, `gohttp` will try to load proxy settings
// from environment variable
func (c *Client) Proxy(proxy string) *Client {
	if proxy != "" {
		c.proxy = proxy
	}
	return c
}

// Cookie set client cookie, the argument is a standard `http.Cookie`
func (c *Client) Cookie(cookie *http.Cookie) *Client {
	if cookie != nil {
		c.cookies = append(c.cookies, cookie)
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

// BasicAuth allows simple username/password authentication
func (c *Client) BasicAuth(username, password string) *Client {
	c.auth.username = username
	c.auth.password = password
	return c
}

// Header sets request header data
func (c *Client) Header(key, value string) *Client {
	c.headers[key] = value
	return c
}

type jsonBodyData struct {
	payload interface{}
}

func (jbd jsonBodyData) Body() (io.Reader, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(jbd.payload)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// JSON accepts a string as data, and sets it as body, and send it as application/json
// If the actual method does not support body or json data, such as `GET`, `HEAD`,
// it will be simply omitted.
func (c *Client) JSON(bodyJSON string) *Client {
	if bodyJSON != "" {
		c.Header(contentType, jsonContentType)
		//TODO: how to handle error
		buf := &bytes.Buffer{}
		buf.WriteString(bodyJSON)
		c.body = buf
	}
	return c
}

// JSONStruct accepts a struct as data, and sets it as body, and send it as application/json
// If the actual method does not support body or json data, such as `GET`, `HEAD`,
// it will be simply omitted.
func (c *Client) JSONStruct(bodyJSON interface{}) *Client {
	if bodyJSON != nil {
		c.Header(contentType, jsonContentType)
		//TODO: how to handle error
		body, _ := jsonBodyData{payload: bodyJSON}.Body()
		c.body = body
	}
	return c
}

type formBodyData struct {
	payload interface{}
}

// Body returns io.Reader from form data.
// Form data is a collection of many key-value pairs,
// so we use go-querystring to parse it to string, then
// create a io.Reader interface.
func (fbd formBodyData) Body() (io.Reader, error) {
	values, err := query.Values(fbd.payload)
	if err != nil {
		return nil, err
	}
	return strings.NewReader(values.Encode()), nil
}

// Form accepts a struct, uses it as body data, and sent it as application/www-x-form-urlencoded
// If the actual method does not support body or form data, such as `GET`, `HEAD`,
// it will be simply omitted.
func (c *Client) Form(bodyForm interface{}) *Client {
	if bodyForm != nil {
		c.Header(contentType, formContentType)
		body, _ := formBodyData{payload: bodyForm}.Body()
		c.body = body
	}
	return c
}

// Body accepts `io.Reader`, will read data from it and use it as request body.
// This doee not set `Content-Type` header, so users should use `Header(key, value)`
// to specify it if necessary.
func (c *Client) Body(body io.Reader) *Client {
	if body != nil {
		c.body = body
	}
	return c
}

// File adds a file in request body, and sends it to server
// Multiple files can be added by calling this method many times
func (c *Client) File(f *os.File, fileName, fieldName string) *Client {
	ff := &fileForm{}
	ff.filename = fileName
	ff.fieldName = fieldName
	ff.file = f

	c.files = append(c.files, ff)
	return c
}

// Get provides a shortcut to send `GET` request
func Get(url string) (*GoResponse, error) {
	return DefaultClient.Get(url)
}

// Head provides a shortcut to send `HEAD` request
func Head(url string) (*GoResponse, error) {
	return DefaultClient.Head(url)
}

// Delete provides a shortcut to send `DELETE` request
// It is used to remove a resource from server
func Delete(url string) (*GoResponse, error) {
	return DefaultClient.Delete(url)
}

// Options provides a shortcut to send `OPTIONS` request
func Options(url string) (*GoResponse, error) {
	return DefaultClient.Options(url)
}

// Post provides a shortcut to send `POST` request
func Post(url string, data io.Reader) (*GoResponse, error) {
	return DefaultClient.Body(data).Post(url)
}

// Put provides a shortcut to send `PUT` request
func Put(url string, data io.Reader) (*GoResponse, error) {
	return DefaultClient.Body(data).Put(url)
}

// Patch provides a shortcut to send `PATCH` request
func Patch(url string, data io.Reader) (*GoResponse, error) {
	return DefaultClient.Body(data).Patch(url)
}
