# gohttp
[![Travis-ci Status](https://travis-ci.org/cizixs/gohttp.svg?branch=master)](https://travis-ci.org/cizixs/gohttp)
[![Coverage Status](https://coveralls.io/repos/github/cizixs/gohttp/badge.svg?branch=master)](https://coveralls.io/github/cizixs/gohttp?branch=master)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square)](LICENSE.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/cizixs/gohttp)](https://goreportcard.com/report/github.com/cizixs/gohttp)
[![GoDoc](https://godoc.org/github.com/cizixs/gohttp?status.svg)](https://godoc.org/github.com/cizixs/gohttp)

A simple to use golang http client

## Features

- [x] All HTTP methods support: GET/HEAD/POST/PUT/PATCH/DELETE/OPTIONS
- [x] Easy to set HTTP Headers
- [x] Query strings can be added through key-value pairs or struct easily
- [x] Extend url path whenever you want
- [x] Send form and json data in one line
- [x] Basic Auth right away
- [x] Read from response the easy way
- [x] Support proxy configuration
- [x] Allow set timeout at all levels
- [x] Automatically retry request, if you let it
- [ ] Custom redirect policy
- [ ] Perform hook functions
- [ ] Session support, persistent response data, and reuse them in next request
- More to come ...

## Principles

- Simple: the exposed interface should be simple and intuitive. After all, this is why `gohttp` is created
- Consistent: Derived from `net/http`, `gohttp` tries to keep consistent with it as much as possible
- Fully-Documented: along with the simple interface, documents and examples will be given to demostrate the full usage, pitfalls and best practices

## Install

```bash
go get github.com/cizixs/gohttp
```

And then import the package in your code:

```go
import "github.com/cizixs/gohttp"
```

## Usage

### Get a resource from url

`gohttp` provides shortcut to make simple `GET` quite straightforward:

    resp, err := gohttp.Get("https://api.github.com/users/cizixs")

The above does the same thing as:

    resp, err := gohttp.New().Get("https://api.github.com/users/cizixs")

In fact, this is exactly what it does behind the scene. 
`gohttp.New()` returns `gohttp.Client` struct, which gives you full control of the request which be sent.

### Create url path on the fly

If url path can not decided until runtime or you want it be flexible, `Path()` method is here to help:

    resp, err := gohttp.New().Path("/repos").Path("cizixs/gohttp/").Path("issues").Get("https://api.github.com/")

Or simply in one method:

    resp, err := gohttp.New().Path("/repos", "cizixs/gohttp/", "issues").Get("https://api.github.com/")

Notice how `gohttp` handles the slash `/` appropriately no matter where it is placed(or not placed at all).

### Pass arguments(query string) in URLs 

There are often times you want to include query strings in url, handling this issue manually can be tiresome and boring.
Not with `gohttp`:

    gohttp.New().Path("/repos", "cizixs/gohttp/", "issues").
        Query("state", "open").Query("sort", "updated").Query("mentioned", "cizixs").
        Get("https://api.github.com/")

Think this is tedious too? Here comes the better part, you can pass a struct as query strings:

    type issueOption struct {
    	State     string `json:"state,omitempty"`
    	Assignee  string `json:"assignee,omitempty"`
    	Creator   string `json:"creator,omitempty"`
    	Mentioned string `json:"mentioned,omitempty"`
    	Labels    string `json:"labels,omitempty"`
    	Sort      string `json:"sort,omitempty"`
    	Rirection string `json:"direction,omitempty"`
    	Since     string `json:"since,omitempty"`
    }

	i := &issueOption{
		State:     "open",
		Mentioned: "cizixs",
		Sort:      "updated",
	}
	resp, err := gohttp.New().Path("/repos", "cizixs/gohttp/", "issues").QueryStruct(i).Get("https://api.github.com/")
    
### Custom Headers

    gohttp.New().Header(key, value).Header(key, value).Get(url)

Or, simply pass all headers in a map:

    gohttp.New().Headers(map[string]string).Get(url)

### Post and PUT

Not only `GET` is simple, `gohttp` implements all other methods:

    body := strings.NewReader("hello, gohttp!")
    gohttp.New().Body(body).Post("https://httpbin.org/post")
    gohttp.New().Body(body).Put("https://httpbin.org/put")

**NOTE**: The actual data sent is based on HTTP method the request finally fires.
Anything not compliant with that METHOD will be ommited.
For example, set data to a `GET` request has no effect, because it will not be used at all.

### Post all kinds of data

When comes to sending data to server, `POST` might be the most frequently used method.
Of all user cases, send form data and send json data comes to the top. 
`gohttp` tries to make these actions easy:

    // send Form data
    gohttp.New().Form("username", "cizixs").Form("password", "secret").Post("https://somesite.com/login")

    // send json data
    gohttp.New().Json(`{"Name":"Cizixs"}`).Post(url)         // use a marshalled json string

    struct User{
        Name string `json:"name,omitempty"`
        Age int `json:"age,omitempty"`
    }

    user := &User{Name: "cizixs", Age: 22}
    gohttp.New().JsonStruct(user).Post(url)   // use a struct and parse it to json

### Basic Auth

```go
gohttp.New().BasicAuth("username", "password").Get("https://api.github.com/users/")
```

### Timeout

By default, `net/http` does not have timeout, will wait forever until response is returned.
This can be a serious issue if server hangs, `gohttp` allows you to set a timeout limit,
if response does not finish in time, an error will be returned.

    gohttp.New().Timeout(100*time.Millisecond).Get("http://example.com")

### Retries

Error happens! `gohttp` provides retry mechanism to automatically resend request when error happens.
This is useful to avoid some unstable error like network temporary failure:

    gohttp.New().Retries(3).Get("http://example.com")

By default, request is only sent only and all. If custom retry is set to less than one, default behavior will be used.
The retry condition only means response error, not including 5XX error.

### Upload file(s)

Upload files is simple too, multiple files can be uploaded in one request.

    f, _ := os.Open(filePath)
    gohttp.New().File(f io.Reader, "filename", "fieldname").Post(url)

### Proxy

If you are sending request behind a proxy, you can do this:

    gohttp.New().Proxy("http://127.0.0.1:4567").Get("http://target.com/cool")

### Cookies

Access cookie from response is simple:

    cookies := resp.Cookies()

    gohttp.New().Cookie(cookie *http.Cookie).Cookie(cookie *http.Cookie).Get(url)

### Response data as string

If the response contains string data, you can read it by:

    resp, _ := gohttp.New().Get("http://someurl.com")
    data, _ := resp.AsString()


### Response data as bytes

If the response contains raw bytes, you can read it by:

    resp, _ := gohttp.New().Get("http://someurl.com")
    data, _ := resp.AsBytes()

### Response data as json struct

If the response contains json struct, you can pass a struct to it, and `gohttp` will marshall it for you:

    user := &User{}
    resp, _ := gohttp.New().Get("http://someurl.com")
    err := resp.AsJSON(user)

### Reuse client

When developing complicated http client applications, it is common to interact with one remote server with some common
settings, like url, basic auth, custom header, etc. `gohttp` provides `New` method for this kind of situation.

As stated in document: 

> New clones current client struct and returns it.
> This is useful to initialize some common parameters and send different requests
> with differenet paths/headers/timeout/retries...

For example:

```go
c := gohttp.New().URL("https://api.github.com/")
c.BasicAuth("cizixs", "mypassword")
c.Timeout(3 * time.Second)
users, err := c.New().Path("/users/").Get()
repos, err := c.New().Path("/repos").Get()
```

**Note**: files, body and cookies value are copied so that if pointer value is used, base client and cloned
client(s) will share the same instance, change on one side will take effect on the other side, and this might not
as expected.

### Debug mode

When developing http apps, it is often necessary to know the actual request and response sent for debugging or testing purpose.
`gohttp` provides debug mode, which can be turned on by environment variable `GOHTTP_DEBUG` or `Debug(bool)` method:

    resp, _ := gohttp.New().Debug(true).Get("http://someurl.com")

In debug mode, `gohttp` will print out each request and response in human-readble format. 

A typical request:

    POST / HTTP/1.1
    Host: 127.0.0.1:43827
    User-Agent: Go-http-client/1.1
    Content-Length: 39
    Content-Type: application/json
    Accept-Encoding: gzip

    {"title":"Test title","name":"cizixs"}

A typical response:

    2016/11/29 18:30:59 HTTP/1.1 200 OK
    Content-Length: 39
    Content-Type: application/json; charset=utf-8
    Date: Tue, 29 Nov 2016 10:30:59 GMT
    
    {"age":24,"name":"cizixs"}

## Contribution

Contributions are welcome!

- Open a new issue if you find a bug or want to propose a feature
- Create a Pull Request if you want to contribute to the code

## Inspiration

This project is heavily influenced by many awesome projects out there, mainly the following:

- [sling: A Go http client library for creating and sending API requests](https://github.com/dghubble/sling)
- [gorequest: simplified HTTP client](https://github.com/parnurzeal/gorequest)
- [requests: HTTP for human](https://github.com/kennethreitz/requests)

## License

[MIT License](https://github.com/cizixs/gohttp/blob/master/LICENSE)
