package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func parseAvailable(z *html.Tokenizer) string {
	for {
		switch z.Next() {
		case html.StartTagToken:
			if z.Token().Data != "b" {
				continue
			}
			if !(z.Next() == html.TextToken && z.Token().Data == "Available:") {
				continue
			}
			tt := z.Next()
			if tt != html.EndTagToken {
				panic(tt)
			}
			tt = z.Next()
			if tt != html.TextToken {
				panic(tt)
			}
			return z.Token().Data
		case html.ErrorToken:
			panic("No available found")
		}
	}
}

func main() {
	code := os.Args[1]
	client := http.Client{Timeout: time.Second * 60}
	data := url.Values{"actions": {"list"}, "code": {code}}
	resp, err := client.PostForm("http://conso.ebox.ca/index.php", data)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic(resp.Status)
	}
	s := strings.Trim(parseAvailable(html.NewTokenizer(resp.Body)), " ")
	i := len(s) - 2
	if s[i:] != " G" {
		panic(s)
	}
	x, err := strconv.ParseFloat(s[:i], 64)
	if err != nil {
		panic(err)
	}
	if x < 1 {
		fmt.Println("Available:", s)
	}
}
