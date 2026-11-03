package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"git.aetherial.dev/aeth/keiji/pkg/auth"
	"git.aetherial.dev/aeth/keiji/pkg/controller"
	"git.aetherial.dev/aeth/keiji/pkg/storage"
	_ "github.com/mattn/go-sqlite3"
)

// newRequest sends an Origin matching the target for unsafe requests. Never use
// a caller-provided Origin or include URL credentials, paths, queries or fragments.
func newRequest(method, target string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, target, body)
	if err != nil {
		return nil, err
	}
	u := req.URL
	if u.Opaque != "" || u.User != nil || u.Hostname() == "" ||
		(!strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https")) {
		return nil, fmt.Errorf("invalid HTTP target URL: %q", target)
	}
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
	default:
		req.Header.Set("Origin", strings.ToLower(u.Scheme)+"://"+u.Host)
	}
	return req, nil
}

// Treat redirects as errors at the call site. In particular, do not replay
// credentials, cookies or mutation bodies at a redirect destination.
func newHTTPClient() *http.Client {
	return &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
}

func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		log.Fatal(err)
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := newRequest(http.MethodPost, uri, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

// authenticate and get the cookie needed to make updates
func authenticate(url, username, password string) *http.Cookie {
	client := newHTTPClient()
	b, _ := json.Marshal(auth.Credentials{Username: username, Password: password})
	req, err := newRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		log.Fatal("auth failed: ", err)
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("auth failed: ", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode > 200 {
		msg, _ := io.ReadAll(resp.Body)
		log.Fatal("Invalid credentials or server error: ", string(msg), "\n Status code: ", resp.StatusCode)
	}
	cookies := resp.Cookies()
	for i := range cookies {
		if cookies[i].Name == controller.AUTH_COOKIE_NAME {
			return cookies[i]
		}
	}
	log.Fatal("Auth cookie not found.")
	return nil
}

// prepare the auth cookie
func prepareCookie(address string) *http.Cookie {

	parsedAddr, err := url.Parse(address)
	if err != nil {
		log.Fatal("unparseable address: ", address, " error: ", err)
	}
	dn := parsedAddr.Hostname()
	var preparedCookie *http.Cookie
	if cookie == "" {
		log.Fatal("Cookie cannot be empty.")
	} else {
		preparedCookie = &http.Cookie{Value: cookie, Name: controller.AUTH_COOKIE_NAME, Domain: dn}
	}
	return preparedCookie
}

var pngFile string
var redirect string
var text string
var col string
var cmd string
var address string
var cookie string

func main() {

	flag.StringVar(&pngFile, "png", "", "The location of the PNG to upload")
	flag.StringVar(&redirect, "redirect", "", "the website that the navbar will redirect to")
	flag.StringVar(&text, "text", "", "the text to display on the menu item")
	flag.StringVar(&col, "col", "", "the column to add/populate the admin table item under")
	flag.StringVar(&cmd, "cmd", "", "the 'command' for the seed program to use, currently supports options 'admin', 'menu', and 'asset', 'nav'")
	flag.StringVar(&address, "address", "https://aetherial.dev", "override the url to contact.")
	flag.StringVar(&cookie, "cookie", "", "pass a cookie to bypass direct authentication")
	flag.Parse()

	client := newHTTPClient()

	switch cmd {
	case "auth":
		cookie := authenticate(fmt.Sprintf("%s/login", address), os.Getenv("KEIJI_USERNAME"), os.Getenv("KEIJI_PASSWORD"))
		fmt.Println(cookie.Value)

	case "asset":
		b, err := os.ReadFile(pngFile)
		if err != nil {
			log.Fatal(err)
		}
		_, fileName := path.Split(pngFile)
		item := storage.Asset{
			Name: fileName,
			Data: b,
		}
		data, _ := json.Marshal(item)
		req, err := newRequest(http.MethodPost, fmt.Sprintf("%s/admin/asset", address), bytes.NewReader(data))
		if err != nil {
			log.Fatal(err)
		}
		req.AddCookie(prepareCookie(address))
		req.Header.Add("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("There was an error performing the desired request: ", err.Error())
			os.Exit(1)
		}
		if resp.StatusCode > 200 {
			defer resp.Body.Close()
			b, _ := io.ReadAll(resp.Body)
			fmt.Println("There was an error performing the desired request: ", string(b))
			os.Exit(2)
		}
		fmt.Println("navigation bar item upload successfully.")
		os.Exit(0)

	case "nav":
		fmt.Println(string(pngFile))
		_, fileName := path.Split(pngFile)
		extraParams := map[string]string{
			"link":     fileName,
			"redirect": redirect,
		}
		req, err := newfileUploadRequest(address+"/admin/navbar", extraParams, "file", pngFile)
		if err != nil {
			log.Fatal(err)
		}
		req.AddCookie(prepareCookie(address))
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("There was an error performing the desired request: ", err.Error())
			os.Exit(1)
		}
		if resp.StatusCode > 200 {
			defer resp.Body.Close()
			b, _ := io.ReadAll(resp.Body)
			fmt.Println("There was an error performing the desired request: ", string(b))
			os.Exit(2)
		}
		fmt.Println("png item upload successfully.")
		os.Exit(0)
	case "menu":
		b, _ := json.Marshal(storage.LinkPair{Text: text, Link: redirect})
		req, err := newRequest(http.MethodPost, fmt.Sprintf("%s/admin/menu", address), bytes.NewReader(b))
		if err != nil {
			log.Fatal(err)
		}
		req.AddCookie(prepareCookie(address))
		req.Header.Add("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("There was an error performing the desired request: ", err.Error())
			os.Exit(1)
		}
		if resp.StatusCode > 200 {
			defer resp.Body.Close()
			b, _ := io.ReadAll(resp.Body)
			fmt.Println("There was an error performing the desired request: ", string(b))
			os.Exit(3)
		}
		fmt.Println("menu item uploaded successfully.")
		os.Exit(0)
	case "admin":
		tables := make(map[string][]storage.TableData)
		adminPage := storage.AdminPage{Tables: tables}
		adminPage.Tables[col] = append(adminPage.Tables[col], storage.TableData{Link: redirect, DisplayName: text})
		b, _ := json.Marshal(adminPage)
		req, err := newRequest(http.MethodPost, fmt.Sprintf("%s/admin/panel", address), bytes.NewReader(b))
		if err != nil {
			log.Fatal(err)
		}
		req.AddCookie(prepareCookie(address))
		req.Header.Add("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("There was an error performing the desired request: ", err.Error())
			os.Exit(1)
		}
		if resp.StatusCode > 200 {
			defer resp.Body.Close()
			b, _ := io.ReadAll(resp.Body)
			fmt.Println("There was an error performing the desired request: ", string(b))
			os.Exit(4)
		}
		fmt.Println("admin item added successfully.")
		os.Exit(0)
	}

}
