// analize-site-to-pdf project main.go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/grokify/html-strip-tags-go" // => strip
)

const READ_URL = "/read/"

type Page struct {
	Text  string
	Table map[string]int
}

func (p *Page) toCount() {
	p.Table = make(map[string]int)
	words := strings.Fields(strip.StripTags(p.Text))

	for i := range words {
		p.Table[words[i]]++
	}
}

func downloadPage(siteUrl string) (*Page, error) {
	resp, err := http.Get("http://" + siteUrl)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &Page{Text: string(body)}, nil
}

func readSiteHandler(w http.ResponseWriter, r *http.Request) {
	siteUrl := r.URL.Path[len(READ_URL):]
	p, _ := downloadPage(siteUrl)

	p.toCount()

	// отправить пдф обратно

	fmt.Fprintf(w, "<h1>Table</h1><div>%s</div>", p.Table)
}

func main() {
	http.HandleFunc(READ_URL, readSiteHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
