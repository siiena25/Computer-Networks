package main

import (
	"bytes"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

const site = "https://news.mail.ru"

func getAttribute(node *html.Node, name string) (string, error) {
	for _, attr := range node.Attr {
		if attr.Key == name {
			return attr.Val, nil
		}
	}
	return "", errors.New("No such attribute")
}

func getClassNames(node *html.Node) []string {
	val, err := getAttribute(node, "class")
	if err != nil {
		return nil
	}
	return strings.Split(val, " ")
}

func hasClassName(node *html.Node, name string) bool {
	names := getClassNames(node)
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

func getElementsByClassName(node *html.Node, name string) []*html.Node {
	var nodes []*html.Node
	if hasClassName(node, name) {
		nodes = append(nodes, node)
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		cn := getElementsByClassName(c, name)
		nodes = append(nodes, cn...)
	}
	return nodes
}

func getElementsByType(node *html.Node, name string) []*html.Node {
	var nodes []*html.Node
	if node.Data == name {
		nodes = append(nodes, node)
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		cn := getElementsByType(c, name)
		nodes = append(nodes, cn...)
	}
	return nodes
}

type item struct {
	Title string
	Preview *preview
	HasPreview bool
	Date string
	Link string
}

type article struct {
	Title string
	Content string
	Url string
}

type preview struct {
	Url string
}

func parsePreview(node *html.Node) *preview {
	imgs := getElementsByType(node, "img")
	src, _ := getAttribute(imgs[0], "data-lazy-block-src")
	return &preview{
		Url: src,
	}
}

func parsePreviewArticle(node *html.Node) string {
	imgs := getElementsByType(node, "img")
	src, _ := getAttribute(imgs[0], "src")
	println(src)
	return src
}

func parseMainPage(document *html.Node) []item {
	blocks := getElementsByClassName(document, "newsitem")
	var res []item
	for _, b := range blocks {
		previewBlock := getElementsByClassName(b, "newsitem__photo")
		var prev *preview
		if len(previewBlock) > 0 {
			prev = parsePreview(previewBlock[0])
		}
		headerBlock := getElementsByClassName(b, "newsitem__title")[0]
		dateBlock := getElementsByClassName(b, "newsitem__param")[0]
		link, _ := getAttribute(headerBlock, "href")
		i := item{
			Title: headerBlock.FirstChild.FirstChild.Data,
			Date: dateBlock.FirstChild.Data,
			Preview: prev,
			HasPreview: previewBlock != nil,
			Link: link[20:],
		}
		res = append(res, i)
	}
	return res
}

func parseBlock(n *html.Node) string {
	if n.FirstChild == nil {
		return n.Data
	}
	res := ""
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		res += parseBlock(c)
	}
	return res
}

func parseArticlePage(document *html.Node) article {
	blocks := getElementsByClassName(document, "article__item_html")
	titleBlock := getElementsByClassName(document, "hdr_collapse")
	previewBlock := getElementsByClassName(document, "photo_article-photo")
	var prev string
	if len(previewBlock) > 0 {
		prev = parsePreviewArticle(previewBlock[0])
	}
	content := ""
	title := ""
	for _, i := range blocks {
		content += parseBlock(i)
	}
	for _, j := range titleBlock {
		title += parseBlock(j)
	}
	return article {
		Content: content,
		Title: title,
		Url: prev,
	}
}

const mainTemplate = `<!DOCTYPE html>
<html>
        <head>
                <title>news.mail.ru</title>
                <meta charset="utf-8" />
        </head>
        <body>
                <h1>news.mail.ru</h1>
                {{ range .items }}
                        <h2> <a href="{{ .Link }}"> {{ .Title }} </a> </h2>
                        {{ if .HasPreview }}
							<img src="{{ .Preview.Url}}" />
                        {{ end }}
                        <br />
                        {{ .Date }}
                {{ end }}
        </body>
</html>
`

const articleTemplate = `<!DOCTYPE html>
<html>
        <head>
                <title>{{ .article.Title }}</title>
                <meta charset="utf-8" />
        </head>
        <body>
                <h1>{{ .article.Title }}</h1>
                {{ .article.Content | html }}
				<br />
				<br />
				<img src="{{ .article.Url}}" />
        </body>
</html>
`


func main() {
	patterns := [4]string{"/society/", "/economics/", "/incident/", "/politics/"}
	for _, i := range patterns {
		http.HandleFunc(i, func(rw http.ResponseWriter, r *http.Request) {
			p, _ := http.Get(site + r.URL.Path)
			doc, _ := html.Parse(p.Body)
			article := parseArticlePage(doc)
			t, _ := template.New("").Parse(articleTemplate)
			rw.WriteHeader(200)
			b := bytes.NewBufferString("")
			_ = t.Execute(b, map[string]interface{}{
				"article": article,
			})
			_, err := rw.Write(b.Bytes())
			if err != nil {
				log.Println(err.Error())
			}
		})
	}

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		p, _ := http.Get(site)
		doc, _ := html.Parse(p.Body)
		items := parseMainPage(doc)
		t, _ := template.New("").Parse(mainTemplate)
		rw.WriteHeader(200)
		b := bytes.NewBufferString("")
		_ = t.Execute(b, map[string]interface{}{
			"items": items,
		})
		_, err := rw.Write(b.Bytes())
		if err != nil {
			log.Println(err.Error())
		}
	})

	_ = http.ListenAndServe(":8016", nil)
}