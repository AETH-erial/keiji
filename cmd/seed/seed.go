package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"git.aetherial.dev/aeth/keiji/pkg/helpers"
	_ "github.com/mattn/go-sqlite3"
)

const DEFAULT_URL = "http://localhost:10277"

// authenticate and get the cookie needed to make updates
func auth() string {
	return ""
}

var pngFile string
var redirect string
var text string
var col string
var cmd string

func main() {

	flag.StringVar(&pngFile, "png", "", "The location of the PNG to upload")
	flag.StringVar(&redirect, "redirect", "", "the website that the navbar will redirect to")
	flag.StringVar(&text, "text", "", "the text to display on the menu item")
	flag.StringVar(&col, "col", "", "the column to add/populate the admin table item under")
	flag.StringVar(&cmd, "cmd", "", "the 'command' for the seed program to use, currently supports options 'admin', 'menu', and 'png")
	flag.Parse()

	client := http.Client{}
	switch cmd {
	case "png":
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
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/admin/navbar", DEFAULT_URL), bytes.NewReader(data))
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
	case "menu":
		b, _ := json.Marshal(helpers.MenuLinkPair{LinkText: text, MenuLink: redirect})
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/admin/menu", DEFAULT_URL), bytes.NewReader(b))
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
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/admin/panel", DEFAULT_URL), bytes.NewReader(b))
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
	}

}
