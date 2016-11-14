# gohttp
[![Travis-ci Status](https://travis-ci.org/cizixs/gohttp.svg?branch=master)](https://travis-ci.org/cizixs/gohttp)
[![Coverage Status](https://coveralls.io/repos/github/cizixs/gohttp/badge.svg?branch=master)](https://coveralls.io/github/cizixs/gohttp?branch=master)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square)](LICENSE.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/cizixs/gohttp)](https://goreportcard.com/report/github.com/cizixs/gohttp)
[![GoDoc](https://godoc.org/github.com/cizixs/gohttp?status.svg)](https://godoc.org/github.com/cizixs/gohttp)

A simple to use golang http client

## Features

- All HTTP methods support: GET/HEAD/POST/PUT/PATCH/DELETE/OPTIONS
- Easy to set HTTP Headers
- Query strings can be added through key-value pairs or struct easily
- Extend url path whenever you want
- Send form and json data in one line

## Principles

- Simple: the exposed interface should be simple and intuitive. After all, this is why `gohttp` is created
- Consistent: Derived from `net/http`, `gohttp` tries to keep consistent with it as much as possible

## Install

```bash
go get github.com/cizixs/gohttp
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

### Upload file(s)

    gohttp.New().File(f Reader).File(f Reader).Post(url)
    gohttp.New().Files(f []Reader).Post(url)

### Redirects

By default, `gohttp` follows redirect automatically, 
If do not want to follow redirect, and return the response as it is:

    gohttp.New().FollowRedict(false).Get()

### Timeout

    gohttp.New().Timeout(timeout).Get()

## Cookies

Access cookie from response is simple:

    cookies := resp.Cookies()

    gohttp.New().Cookie(key, value string).Cookie(key, value string).Get(url)
    gohttp.New().Cookies(map[string]string).Get(url)

For more control, use CookieJar(specify cookie idle time, host, url etc):

    gohttp.New().CookieJar(cj).Get()

## Errors and Exceptions

TODO:
When to raise errors, and how many kind of errors.

## Session

Client gives you flexibility on how to sent each request, which `session` helps to
persistent certain parameters across requests.

The most common usage would be persisting cookies, making it possible to handle login.

    s := gohttp.NewSession()
    s.Form(user, pass).Post(url)
    s.Get(url)

All parameters will be merged with existing ones(if merge is supported, like `header`) or overided(like `timeout`) by.

## HTTPS

TODO

## Authentication

### Basic Auth

```go
gohttp.New().BasicAuth("username", "password").Get("https://api.github.com/users/")
```

## Proxies

TODO

## Hooks

Hooks let you perform some actions at certain checkpoint.
