package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/publicsuffix"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

type credential struct {
	Usrname string
	Pwd     string
}

func (c *credential) parse(path string) {
	file, err := os.Open(path)
	panicOnError(err)
	defer file.Close()
	panicOnError(json.NewDecoder(file).Decode(c))
}

func csrf(client *http.Client) string {
	resp, err := client.Get("https://client.ebox.ca/")
	panicOnError(err)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic(resp.Status)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	panicOnError(err)
	ans, exists := doc.Find("input[name='_csrf_security_token']").Attr("value")
	if !exists {
		panic("Fail to find _csrf_security_token")
	}
	return ans
}

func login(client *http.Client, cred *credential) {
	data := url.Values{
		"usrname": {cred.Usrname},
		"pwd":     {cred.Pwd},
		"_csrf_security_token": {csrf(client)},
	}
	resp, err := client.PostForm("https://client.ebox.ca/login", data)
	panicOnError(err)
	defer resp.Body.Close()
	if resp.StatusCode != 302 {
		log.Panicf("Login failed, return %v", resp.Status)
	}
	url, err := resp.Location()
	panicOnError(err)
	if url.Path != "/home" {
		log.Panicf("Login failed, redirect to %v", url.Path)
	}
}

func parseFloat(s string) float64 {
	ans, err := strconv.ParseFloat(s, 64)
	panicOnError(err)
	return ans
}

func usage(client *http.Client) {
	resp, err := client.Get("https://client.ebox.ca/myusage")
	panicOnError(err)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic(resp.Status)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	panicOnError(err)
	text := doc.Find(".usage_summary").Children().First().Text()
	fields := strings.Fields(text)
	if len(fields) != 5 || fields[1] != "Go" || fields[2] != "/" || fields[4] != "Go" {
		log.Panicf("Invalid text: %v", text)
	}
	if parseFloat(fields[0]) / parseFloat(fields[3]) > 0.9 {
		fmt.Printf("EBOX Usage: %v / %v\n", fields[0], fields[3])
	}
}

func main() {
	var cred credential
	cred.parse(os.Args[1])
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	panicOnError(err)
	client := http.Client{
		Timeout: time.Second * 60,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	login(&client, &cred)
	usage(&client)
}
