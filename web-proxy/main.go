package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"golang.org/x/net/html"
)

const site = "http://localhost"
const port = ":8888"
const feed = "/site/"
const path = site + port + feed

func getAttrValue(node *html.Node, attrName string) string {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}
	return ""
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func changeURL(node *html.Node, domen string, ind int) {
	if node.Data == "head" {
		base := &html.Node{Data: "base", Type: html.ElementNode}
		base.Attr = append(base.Attr, html.Attribute{Key: "href", Val: path + domen})
		node.AppendChild(base)
	}

	if node.Data == "meta" && getAttrValue(node, "http-equiv") == "Content-Security-Policy" {
		node.Data = "safe"
		return
	}

	for i, attr := range node.Attr {
		if attr.Key == "href" || attr.Key == "src" {
			attr.Val = strings.TrimSpace(attr.Val)
			urlParse, err := url.Parse(attr.Val)
			if err == nil {
				if urlParse.IsAbs() {
					attr.Val = path + attr.Val[strings.Index(attr.Val, "://")+3:]
				} else if len(attr.Val) > 2 && attr.Val[1] == '/' {
					attr.Val = path + attr.Val[2:]
				} else if len(attr.Val) > 1 && attr.Val[0] == '/' {
					attr.Val = path + domen + attr.Val
				}
			}
			node.Attr[i].Val = attr.Val
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		changeURL(child, domen, ind + 1)
	}
}

func doSite(url string, originalResp http.ResponseWriter) bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}

	contentType := resp.Header.Get("Content-type")
	copyHeader(originalResp.Header(), resp.Header)
	originalResp.WriteHeader(resp.StatusCode)

	if strings.Contains(contentType, "text/html") {
		node, err := html.Parse(resp.Body)

		if err != nil {
			log.Println(err.Error())
			return false
		}

		domen := strings.Split(url, "/")[2]
		changeURL(node, domen, 0)

		var buffer bytes.Buffer
		html.Render(&buffer, node)
		ans := buffer.String()
		fmt.Fprint(originalResp, ans)
	} else {
		io.Copy(originalResp, resp.Body)
	}
	return true
}

type protocol struct {
	Protocol string
}

func main() {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		log.Fatal(err.Error())
	}
	protocols := []protocol{{"http://"}, {"https://"}}

	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		if err := tmpl.Execute(resp, map[string]interface{}{
			"protocols": protocols,
			"path":  path,
		}); err != nil {
			log.Fatal(err.Error())
		}
	})
	http.HandleFunc(feed, func(resp http.ResponseWriter, req *http.Request) {
		for i, protocol := range protocols {
			page := protocol.Protocol + req.URL.String()[len(feed):]
			log.Println(page, i)
			if doSite(page, resp) {
				break
			}
		}
	})
	http.ListenAndServe(port, nil)
}