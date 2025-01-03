package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"

	"git.aetherial.dev/aeth/keiji/pkg/controller"
	"git.aetherial.dev/aeth/keiji/pkg/helpers"
	_ "github.com/mattn/go-sqlite3"
)

// authenticate and get the cookie needed to make updates
func auth(url, username, password string) *http.Cookie {
	client := http.Client{}
	b, _ := json.Marshal(helpers.Credentials{Username: username, Password: password})
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
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
	dn := parsedAddr.Hostname()
	if err != nil {
		log.Fatal("unparseable address: ", address, " error: ", err)
	}
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

	client := http.Client{}

	switch cmd {
	case "auth":
		cookie := auth(fmt.Sprintf("%s/login", address), os.Getenv("KEIJI_USERNAME"), os.Getenv("KEIJI_PASSWORD"))
		fmt.Println(cookie.Value)

	case "asset":
		b, err := os.ReadFile(pngFile)
		if err != nil {
			log.Fatal(err)
		}
		_, fileName := path.Split(pngFile)
		item := helpers.Asset{
			Name: fileName,
			Data: b,
		}
		data, _ := json.Marshal(item)
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/admin/asset", address), bytes.NewReader(data))
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
		b, err := os.ReadFile(pngFile)
		if err != nil {
			log.Fatal(err)
		}
		_, fileName := path.Split(pngFile)
		fmt.Println(fileName)
		item := helpers.NavBarItem{
			Link:     fileName,
			Redirect: redirect,
			Png:      b,
		}
		data, _ := json.Marshal(item)
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/admin/navbar", address), bytes.NewReader(data))
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
		fmt.Println("png item upload successfully.")
		os.Exit(0)
	case "menu":
		b, _ := json.Marshal(helpers.MenuLinkPair{LinkText: text, MenuLink: redirect})
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/admin/menu", address), bytes.NewReader(b))
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
		tables := make(map[string][]helpers.TableData)
		adminPage := helpers.AdminPage{Tables: tables}
		adminPage.Tables[col] = append(adminPage.Tables[col], helpers.TableData{Link: redirect, DisplayName: text})
		b, _ := json.Marshal(adminPage)
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/admin/panel", address), bytes.NewReader(b))
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
