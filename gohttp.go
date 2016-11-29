package gohttp

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
)

const (
	contentType     = "Content-Type"
	jsonContentType = "application/json"
	formContentType = "application/x-www-form-urlencoded"
)

// DefaultTimeout defines the request timeout limit, avoiding client hanging when the
// remote server does not responde.
const DefaultTimeout = 3 * time.Second

// debugEnv is the environment variable that toggles debug mode.
// 1 means turn on debug mode, 0 or any other value(empty value) means turn off debug mode
const debugEnv = "GOHTTP_DEBUG"

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
// redirect history etc.
type GoResponse struct {
	*http.Response
}

// AsString returns the response data as string
// An error wil bw returned if the body can not be read as string
func (resp *GoResponse) AsString() (string, error) {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// AsBytes return the response body as byte slice
// An error will be returned if the body can not be read as bytes
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

	// timeout sets the waiting time before request is finished
	// If request exceeds the time, error will be returned.
	// The default value zero means no timeout, which is what the `net/http` DefaultClient does.
	timeout time.Duration

	// TLSHandshakeTimeout limits the time spent performing TLS handshake
	tlsHandshakeTimeout time.Duration

	// how many attempts will be used before give up on error
	retries int

	// transport is the actual worker that carries http request, and send it out.
	transport *http.Transport

	// debug toggles debug mode of gohttp.
	// It is useful when user wants to see what is going on behind the scene.
	debug bool
}

// DefaultClient provides a simple usable client, it is given for quick usage.
// For more control, please create a client manually.
var DefaultClient = New()

// New returns a new GoClient with default values
func New() *Client {
	t := &http.Transport{}
	debug := os.Getenv(debugEnv) == "1"

	return &Client{
		query:        make(map[string]string),
		queryStructs: make([]interface{}, 0),
		headers:      make(map[string]string),
		auth:         basicAuth{},
		cookies:      make([]*http.Cookie, 0),
		files:        make([]*fileForm, 0),
		timeout:      DefaultTimeout,
		transport:    t,
		debug:        debug,
	}
}

func copyMap(src map[string]string) map[string]string {
	newMap := make(map[string]string)
	for k, v := range src {
		newMap[k] = v
	}

	return newMap
}

// New clones current client struct and returns it
// This is useful to initialize some common parameters and send different requests
// with differenet paths/headers/...
//
// For example:
//    c := gohttp.New().URL("https://api.github.com/")
//    c.BasicAuth("cizixs", "mypassword")
//    users, err := c.New().Path("/users/").Get()
//    repos, err := c.New().Path("/repos").Get()
//
// Note that files, body and cookies value are copied if pointer value is used, base client and cloned
// client(s) will share the same instance, change on one side will take effect on the other side.
func (c *Client) New() *Client {
	newClient := &Client{}
	// simple copy
	newClient.url = c.url
	newClient.path = c.path
	newClient.auth = c.auth
	newClient.proxy = c.proxy
	newClient.timeout = c.timeout
	newClient.tlsHandshakeTimeout = c.tlsHandshakeTimeout
	newClient.retries = c.retries
	newClient.debug = c.debug

	// make a copy of simple map data
	// NOTE: if the map data contains pointer value, it will be shallow copy.
	// Right now it works, because both headers and queries are `string-string` map value
	newClient.query = copyMap(c.query)
	newClient.headers = copyMap(c.headers)

	// TODO(cizixs): maybe make a deep copy for these pointer values, or just leave them out?
	newClient.c = c.c
	newClient.queryStructs = c.queryStructs
	newClient.body = c.body
	newClient.cookies = c.cookies
	newClient.files = c.files

	// use the same tranport
	newClient.transport = c.transport

	return newClient
}

// setupClient handles the connection details from http client to TCP connections.
// Timeout, proxy, TLS config ..., these are very important but rarely used directly
// by httpclient users.
func (c *Client) setupClient() error {
	// create the transport and client instance first
	if c.proxy != "" {
		// use passed proxy, otherwise try to use environment variable proxy, or just no proxy at all.
		proxy, err := url.Parse(c.proxy)
		if err != nil {
			return err
		}
		c.transport.Proxy = http.ProxyURL(proxy)
	}

	if c.tlsHandshakeTimeout != time.Duration(0) {
		c.transport.TLSHandshakeTimeout = c.tlsHandshakeTimeout
	}

	// TODO(cizixs): maybe reuse http.Client as well
	c.c = &http.Client{Transport: c.transport}

	// request timeout limit
	// timeout zero means no timeout
	// NOTE: `Timeout` property is only supported from go1.3+
	if c.timeout != time.Duration(0) {
		c.c.Timeout = c.timeout
	}
	return nil
}

func (c *Client) prepareFiles() error {
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
				return err
			}
			_, err = io.Copy(part, file.file)
			if err != nil {
				return err
			}
		}

		err := writer.Close()
		if err != nil {
			return err
		}

		// finally, get the real content type, and set the header
		multipartContentType := writer.FormDataContentType()
		c.Header(contentType, multipartContentType)
		c.body = body
	}

	return nil
}

// prepareRequest does all the preparation jobs for `gohttp`.
// The main job is create and configure all structs like `Transport`, `Dialer`, `Client`
// according to arguments passed to `gohttp`.
// TODO(cizixs): This method is getting longer and longer, will try to tidy it up, and
// move some content to individual functions.
func (c *Client) prepareRequest(method string) (*http.Request, error) {
	err := c.setupClient()
	if err != nil {
		return nil, err
	}

	err = c.prepareFiles()
	if err != nil {
		return nil, err
	}

	// create the basic request
	req, err := http.NewRequest(method, c.url, c.body)

	// concatenate path to url if exists
	if c.path != "" {
		// Adds prefix "/" if necessary
		if !strings.HasPrefix(c.path, "/") {
			c.path = "/" + c.path
		}
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

// Debug sets the debug mode of go http.
// By default, debug mode is off.
//
// This is only for testing and debugging purpose,
// it will print out detail information such as request and response dump string,
// and other logging lines.
//
// If `GOHTTP_DEBUG` environment variable is set, `gohttp` will use the value.
// 1 for turn on debug mode, and others for turn off debug mode.
// Debug method overrides environment variable value.
func (c *Client) Debug(debug bool) *Client {
	c.debug = debug
	return c
}

// Do takes HTTP and url, then makes the request, return the response.
// All other HTTP methods will call `Do` behind the scene, and
// it can be used directly to send the request.
// Custom HTTP method can be sent with this method.
// Accept optional url parameter, if multiple urls are given only the first one
// will be used.
func (c *Client) Do(method string, urls ...string) (*GoResponse, error) {
	// TODO: check if url is valid
	url := ""
	if len(urls) >= 1 && urls[0] != "" {
		url = urls[0]
	}
	c.URL(url)

	req, err := c.prepareRequest(method)
	if err != nil {
		return nil, err
	}

	// use httputil to dump raw request string.
	// NOTE: some details might be lost such as header order and case.
	if c.debug {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			log.Println("err: ", err)
			return nil, err
		}
		log.Printf("%s\n", string(dump))
	}

	var resp *http.Response
	// retry the request certain time, if error happens
	tried := 0
	for {
		resp, err = c.c.Do(req)
		tried++
		if c.retries <= 1 || tried >= c.retries || err == nil {
			break
		} else {
			log.Printf("Request [%d/%d] error: %v, retrying...\n", tried, c.retries, err)
		}
	}
	if err != nil {
		log.Printf("Final request error after %d attempt(s): %v\n", tried, err)
		return nil, err
	}

	if c.debug {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Println("err: ", err)
			return nil, err
		}
		log.Printf("%s\n", string(dump))
	}
	return &GoResponse{resp}, err
}

// Get handles HTTP GET request, and return response to user
// Note that the response is not `http.Response`, but a thin wrapper which does
// exactly what it used to and a little more.
func (c *Client) Get(urls ...string) (*GoResponse, error) {
	return c.Do("GET", urls...)
}

// Post handles HTTP POST request
func (c *Client) Post(urls ...string) (*GoResponse, error) {
	return c.Do("POST", urls...)
}

// Head handles HTTP HEAD request
// HEAD request works the same way as GET, except the response body is empty.
func (c *Client) Head(urls ...string) (*GoResponse, error) {
	return c.Do("HEAD", urls...)
}

// Put handles HTTP PUT request
func (c *Client) Put(urls ...string) (*GoResponse, error) {
	return c.Do("PUT", urls...)
}

// Delete handles HTTP DELETE request
func (c *Client) Delete(urls ...string) (*GoResponse, error) {
	return c.Do("DELETE", urls...)
}

// Patch handles HTTP PATCH request
func (c *Client) Patch(urls ...string) (*GoResponse, error) {
	return c.Do("PATCH", urls...)
}

// Options handles HTTP OPTIONS request
func (c *Client) Options(urls ...string) (*GoResponse, error) {
	return c.Do("OPTIONS", urls...)
}

// URL sets the base url for the client, this can be overrided by the parameter of the final
// GET/POST/PUT method
func (c *Client) URL(url string) *Client {
	if url != "" {
		c.url = url
	}
	return c
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
//
// Usage:
//    gohttp.New().Proxy("http://127.0.0.1:4567").Get("http://someurl.com")
func (c *Client) Proxy(proxy string) *Client {
	if proxy != "" {
		c.proxy = proxy
	}
	return c
}

// Timeout sets the wait limit for a request to finish.
// This time includes connection time, redirectoin time, and
// read response body time. If request does not finish before the timeout,
// any ongoing action will be interrupted and an error will return
//
// Usage:
//    gohttp.New().Timeout(time.Second * 10).Get(url)
func (c *Client) Timeout(timeout time.Duration) *Client {
	// TODO(cizixs): add other timeouts like setup connection timeout,
	c.timeout = timeout
	return c
}

// TLSHandshakeTimeout sets the wait limit for performing TLS handshake.
// Default value for go1.6 is 10s, change this value to suit yourself.
func (c *Client) TLSHandshakeTimeout(timeout time.Duration) *Client {
	c.tlsHandshakeTimeout = timeout
	return c
}

// Retries set how many request attempts will be conducted if error happens for a request.
// number <= 1 means no retries, send one request and finish.
func (c *Client) Retries(n int) *Client {
	// TODO(cizixs): allow user to customize retry condition, like if the response status code is 5XX.
	c.retries = n
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
