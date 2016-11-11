# gohttp
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square)](LICENSE.md)

A simple to use golang http client

## QuickStart

### Get a resource from url

    gohttp.Get("api.github.com/v2/users/cizixs")
    gohttp.HEAD("api.github.com/v2/users/cizixs")

### Pass arguments(query string) in URLs 

    gohttp.New().Query(key, value).Query(key, value).Get(url)

### Post and PUT

    gohttp.New().Data(data).Post(url)
    gohttp.New().Data(data).Put(url)

**NOTE**: The actual data sent is based on HTTP method the request finally fires, 
anything not compliant with that METHOD will be ommited if it does not cause any error.
For example, set data to a `GET` request has no effect, because it will not be used at all.

### Response Data Streaming

### Custom Headers

    gohttp.New().Header(key, value).Header(key, value).Get(url)

Or, simply pass all headers in a map:

    gohttp.New().Headers(map[string]string).Get(url)

### Post all kinds of data

    g := gohttp.New()
    // send form data, key=value pairs
    g.Form(key, value).Form(key, value).Post(url)

    // send json data
    g.Json(string).Post(url)         // use a marshalled json string
    g.JsonStruct(struct).Post(url)   // use a struct and parse it to json

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

## Shortcuts

### Handle json response

    var user User
    gohttp.New().Json(url, &user)

## Authentication

TODO

## Proxies

TODO

## Hooks

Hooks let you perform some actions at certain checkpoint.
