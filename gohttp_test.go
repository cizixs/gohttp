package gohttp_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestResponseAsString(t *testing.T) {
	assert := assert.New(t)

	greeting := "hello, gohttp."
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, greeting)
	}))
	defer ts.Close()

	resp, err := gohttp.Get(ts.URL)
	assert.NoError(err, "A get request should cause no error.")
	assert.Equal(http.StatusOK, resp.StatusCode)

	actualGreeting, err := resp.AsString()
	assert.NoError(err, "read the response body should not cause error.")
	assert.Equal(greeting, actualGreeting)
}

func TestResponseAsBytes(t *testing.T) {
	assert := assert.New(t)

	greeting := "hello, gohttp."
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, greeting)
	}))
	defer ts.Close()

	resp, err := gohttp.Get(ts.URL)
	assert.NoError(err, "A get request should cause no error.")
	assert.Equal(http.StatusOK, resp.StatusCode)

	actualGreeting, err := resp.AsBytes()
	assert.NoError(err, "read the response body should not cause error.")
	assert.Equal([]byte(greeting), actualGreeting)
}

func TestResponseAsJSON(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(w, r.Body)
	}))
	defer ts.Close()

	resp, err := gohttp.New().JSON(`{"Name":"gohttp"}`).Post(ts.URL)
	assert.NoError(err, "A get request should cause no error.")
	assert.Equal(http.StatusOK, resp.StatusCode)

	repo := &struct {
		Name string `json:"name,omitempty"`
	}{}
	err = resp.AsJSON(repo)
	assert.NoError(err, "read the response body should not cause error.")
	assert.Equal("gohttp", repo.Name)
}

func TestHead(t *testing.T) {
	assert := assert.New(t)

	testHeader := "Test-Header"
	// test server that writes HTTP method back
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Because HEAD response does not include body, we set the received method in header
		w.Header().Add(testHeader, r.Method)
	}))
	defer ts.Close()

	resp, _ := gohttp.Head(ts.URL)
	assert.Equal("HEAD", resp.Header.Get(testHeader))
}

func TestDelete(t *testing.T) {
	assert := assert.New(t)

	// test server that writes HTTP method back
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.Method))
	}))
	defer ts.Close()

	resp, _ := gohttp.Delete(ts.URL)
	method, _ := ioutil.ReadAll(resp.Body)
	assert.Equal("DELETE", string(method))
}

func TestPost(t *testing.T) {
	assert := assert.New(t)

	// test server that writes HTTP method back
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.Method))
	}))
	defer ts.Close()

	resp, _ := gohttp.Post(ts.URL, nil)
	method, _ := ioutil.ReadAll(resp.Body)
	assert.Equal("POST", string(method))
}

func TestPatch(t *testing.T) {
	assert := assert.New(t)

	// test server that writes HTTP method back
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.Method))
	}))
	defer ts.Close()

	resp, _ := gohttp.Patch(ts.URL, nil)
	method, _ := ioutil.ReadAll(resp.Body)
	assert.Equal("PATCH", string(method))
}

func TestOptions(t *testing.T) {
	assert := assert.New(t)

	// test server that writes HTTP method back
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.Method))
	}))
	defer ts.Close()

	resp, _ := gohttp.Options(ts.URL)
	method, _ := ioutil.ReadAll(resp.Body)
	assert.Equal("OPTIONS", string(method))
}

func TestPut(t *testing.T) {
	assert := assert.New(t)

	// test server that writes HTTP method back
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.Method))
	}))
	defer ts.Close()

	resp, _ := gohttp.Put(ts.URL, nil)
	method, _ := ioutil.ReadAll(resp.Body)
	assert.Equal("PUT", string(method))
}

func TestCookie(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%v", r.Cookies())
	}))
	defer ts.Close()

	cookie := &http.Cookie{Name: "foo", Value: "bar"}

	resp, err := gohttp.New().Cookie(cookie).Get(ts.URL)
	assert.NoError(err, "Get request with cookie should cause no error.")
	data, _ := ioutil.ReadAll(resp.Body)
	assert.Equal("[foo=bar]", string(data))
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

	resp, err := gohttp.New().JSON(`{"Name":"Cizixs"}`).Post(ts.URL)
	assert.NoError(err, "Post request should cause no error.")
	data, _ := ioutil.ReadAll(resp.Body)
	returnedUser := User{}
	json.Unmarshal(data, &returnedUser)

	assert.Equal("Cizixs", returnedUser.Name)
	assert.Equal(0, returnedUser.Age)
}

func TestPostJSONStruct(t *testing.T) {
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

	resp, err := gohttp.New().JSONStruct(user).Post(ts.URL)
	assert.NoError(err, "Post request should cause no error.")
	data, _ := ioutil.ReadAll(resp.Body)
	returnedUser := User{}
	json.Unmarshal(data, &returnedUser)

	assert.Equal("Test title", returnedUser.Title)
	assert.Equal("cizixs", returnedUser.Name)
	assert.Equal(0, returnedUser.Age)
}

func TestPostFiles(t *testing.T) {
	assert := assert.New(t)

	// the mock http server will parse files uploaded, and send `filename:fileLength` string back to client
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reader, err := r.MultipartReader()
		if err != nil {
			fmt.Fprintf(w, "Ops: %v\n", err)
		} else {
			files := []string{}
			for {
				part, err := reader.NextPart()
				if err == io.EOF || part == nil {
					break
				}
				if err != nil {
					continue
				}
				if part.FileName() == "" {
					fmt.Fprintf(w, "Ops: empty file name\n")
					continue
				}
				data, err := ioutil.ReadAll(part)
				if err != nil {
					fmt.Fprintf(w, "Ops: %v\n", err)
				}
				files = append(files, fmt.Sprintf("%s:%d", part.FileName(), len(string(data))))
			}
			fmt.Fprintf(w, strings.Join(files, "&"))
		}
	}))

	filename := "./LICENSE"
	f, _ := os.Open(filename)
	resp, _ := gohttp.New().File(f, "hello.txt", "myfiled").Post(ts.URL)
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Ops-Ops: %v\n", err)
	}

	assert.Equal("hello.txt:1063", string(data))
}
