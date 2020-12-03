package main

import (
	"github.com/mgutz/logxi/v1"
	"golang.org/x/net/html"
	"net/http"
	"io"
	"strconv"
)

func getChildren(node *html.Node) []*html.Node {
	var children []*html.Node
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		children = append(children, c)
	}
	return children
}

func getAttr(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func isText(node *html.Node) bool {
	return node != nil && node.Type == html.TextNode
}

func isElem(node *html.Node, tag string) bool {
	return node != nil && node.Type == html.ElementNode && node.Data == tag
}

func isDiv(node *html.Node, class string) bool {
	return isElem(node, "div") && getAttr(node, "class") == class
}

func readItem(item *html.Node) *Item {
	log.Info("<---in readItem--->")
	if a := item.FirstChild; isElem(a, "a") {
		cs := getChildren(a)
		log.Info("len(cs)", "val", len(cs))
		log.Info("cs[0].Data", "val", cs[0].Data)
		if (len(cs) >= 2){
			log.Info("cs[1].Data", "val", cs[1].Data)
		}
		log.Info("<a href=\"val\">", "val", getAttr(a, "href"))

		if (len(cs) == 1){
			log.Info("<---out readItem--->")
			return &Item{
				Ref:   getAttr(a, "href"),
				Time:  "no information about time",
				Title: cs[0].Data,
			}
		} else {
			log.Info("<---out readItem--->")
			return &Item{
				Ref:   getAttr(a, "href"),
				Time:  getAttr(cs[0], "title"),
				Title: cs[1].Data,
			}
		}
	}
	return nil
}

type Item struct {
	Ref, Time, Title string
}


func downloadNews() []*Item {
	log.Info("sending request to lenta.ru")
	if response, err := http.Get("http://lenta.ru"); err != nil {
		log.Error("request to lenta.ru failed", "error", err)
	} else {
		defer response.Body.Close()
		status := response.StatusCode
		log.Info("got response from lenta.ru", "status", status)
		if status == http.StatusOK {
			if doc, err := html.Parse(response.Body); err != nil {
				log.Error("invalid HTML from lenta.ru", "error", err)
			} else {
				log.Info("HTML from lenta.ru parsed successfully")
				return search(doc)
			}
		}
	}
	return nil
}

func search(node *html.Node) []*Item {
	var allItems []*Item
	if isDiv(node, "b-yellow-box__wrap") || isDiv(node, "span4") {
		var items []*Item
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if isDiv(c, "item") {
				if item := readItem(c); item != nil {
					items = append(items, item)
				}
			} else if isDiv(c, "first-item"){
				i := node.FirstChild
				i = i.NextSibling
				if isDiv(i, "h2"){
					if item := readItem(i); item != nil {
						items = append(items, item)
					}
				}
			}
		}
		return items
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if items := search(c); items != nil {
			allItems = append(allItems, items ...)
		}
	}
	return allItems
}

func Router(w http.ResponseWriter, r *http.Request) {
	count := 1
	for _, item := range items{
		href := "http://lenta.ru" + item.Ref
		io.WriteString(w, "<p>______________________________________________________________________________________________________________</p>")
		io.WriteString(w, "<br>№" + strconv.Itoa(count) + "<br>")
		io.WriteString(w, "Заголовок: " + item.Title + "<br>")
		io.WriteString(w, "Дата: " + item.Time + "<br>")
		io.WriteString(w, "<a href=\"" + href + "\">Ссылка</a>" + "<br>")
		count++
	}
}

var items []*Item

func main() {
	println("start")
	log.Info("Downloader started")
	items = downloadNews()
	http.HandleFunc("/", Router)
	err := http.ListenAndServe(":3015", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}