// analize-site-to-pdf project main.go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"strconv"
	"strings"
	"unicode"

	"github.com/grokify/html-strip-tags-go" // => strip
)

const READ_URL = "/read/"

type Page struct {
	Text  string
	Table map[string]int
}

func (p *Page) toCount() {
	f := func(c rune) bool {
		return !unicode.IsLetter(c) // && !unicode.IsNumber(c)
	}
	words := strings.FieldsFunc(strip.StripTags(p.Text), f)
	p.Table = make(map[string]int)

	for i := range words {
		p.Table[strings.ToLower(words[i])]++
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

	minOccurrence := 1 // default minimal word occurrence
	m, _ := url.ParseQuery(r.URL.RawQuery)
	if m != nil {
		minOccurrence, _ = strconv.Atoi(m["occ"][0])
		if minOccurrence <= 0 {
			minOccurrence = 1
		}
	}

	// отправить пдф обратно

	fmt.Fprintf(w, "<h1>Table</h1><div>%s</div>", p.Table)
}

func main() {
	http.HandleFunc(READ_URL, readSiteHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
