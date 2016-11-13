package gohttp_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cizixs/gohttp"
)

func TestGet(t *testing.T) {
	assert := assert.New(t)

	greeting := "hello, gohttp."
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, greeting)
	}))
	defer ts.Close()

	resp, err := gohttp.Get(ts.URL)
	assert.NoError(err, "A get request should cause no error.")
	assert.Equal(http.StatusOK, resp.StatusCode)

	actualGreeting, err := ioutil.ReadAll(resp.Body)
	assert.Equal(greeting, string(actualGreeting))
}

func TestGetWithPath(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.URL.Path)
	}))
	defer ts.Close()

	resp, err := gohttp.New().Path("/users").Path("cizixs").Get(ts.URL)
	assert.NoError(err, "Get request with path should cause no error.")
	data, _ := ioutil.ReadAll(resp.Body)
	assert.Equal("/users/cizixs", string(data))
}

func TestGetWithQuery(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, r.URL.RawQuery)
	}))
	defer ts.Close()

	resp, err := gohttp.New().Query("foo", "bar").Get(ts.URL)
	assert.NoError(err, "Get request with query string should cause no error.")
	data, _ := ioutil.ReadAll(resp.Body)
	assert.Equal("foo=bar", string(data))

	resp, err = gohttp.New().Query("foo", "bar").Query("name", "cizixs").Get(ts.URL)
	assert.NoError(err, "Get request with query string should cause no error.")
	data, _ = ioutil.ReadAll(resp.Body)
	assert.Equal("foo=bar&name=cizixs", string(data))
}

func TestGetWithQueryStruct(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, r.URL.RawQuery)
	}))
	defer ts.Close()

	options := struct {
		Query   string `url:"q"`
		ShowAll bool   `url:"all"`
		Page    int    `url:"page"`
	}{"foo", true, 2}

	resp, err := gohttp.New().QueryStruct(options).Get(ts.URL)
	assert.NoError(err, "Get request with query string should cause no error.")
	data, _ := ioutil.ReadAll(resp.Body)
	assert.Equal("all=true&page=2&q=foo", string(data))
}

func TestGetWithHeader(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Write(w)
	}))
	defer ts.Close()

	userAgent := "gohttp client by cizixs"
	resp, err := gohttp.New().Header("User-Agent", userAgent).Get(ts.URL)
	assert.NoError(err, "Get request with header should cause no error.")
	data, _ := ioutil.ReadAll(resp.Body)
	assert.True(strings.Contains(string(data), "User-Agent"))
	assert.True(strings.Contains(string(data), userAgent))
}

func TestPostForm(t *testing.T) {
	assert := assert.New(t)

	type Login struct {
		Name     string `json:"name,omitempty"`
		Password string `json:"password,omitempty"`
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := ioutil.ReadAll(r.Body)
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			w.Write([]byte("No form data"))
		} else {
			fmt.Fprint(w, string(data))
		}
	}))
	defer ts.Close()

	user := Login{
		Name:     "cizixs",
		Password: "test1234",
	}

	resp, err := gohttp.New().Form(user).Post(ts.URL)
	assert.NoError(err, "Post request should cause no error.")
	data, _ := ioutil.ReadAll(resp.Body)

	assert.Equal("Name=cizixs&Password=test1234", string(data))
}

func TestPostJSON(t *testing.T) {
	assert := assert.New(t)

	type User struct {
		Title string `json:"title,omitempty"`
		Name  string `json:"name,omitempty"`
		Age   int    `json:"age,omitempty"`
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := ioutil.ReadAll(r.Body)
		if r.Header.Get("Content-Type") != "application/json" {
			w.Write([]byte("No json data"))
		} else {
			fmt.Fprint(w, string(data))
		}
	}))
	defer ts.Close()

	user := User{
		Title: "Test title",
		Name:  "cizixs",
	}

	resp, err := gohttp.New().JSON(user).Post(ts.URL)
	assert.NoError(err, "Post request should cause no error.")
	data, _ := ioutil.ReadAll(resp.Body)
	returnedUser := User{}
	json.Unmarshal(data, &returnedUser)

	assert.Equal("Test title", returnedUser.Title)
	assert.Equal("cizixs", returnedUser.Name)
	assert.Equal(0, returnedUser.Age)
}
