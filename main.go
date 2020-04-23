// analize-site-to-pdf project main.go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"sort"
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

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

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

func rankByWordCount(wordFrequencies map[string]int) PairList {
	pl := make(PairList, len(wordFrequencies))
	i := 0
	for k, v := range wordFrequencies {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
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

	sortedPairs := rankByWordCount(p.Table)
	// TODO отправить пдф обратно

	fmt.Fprintf(w, "<h1>Table</h1><table>")

	for count, pair := range sortedPairs {
		fmt.Fprintf(w, "<tr><td>%d</td><td>%s</td></tr>", pair.Value, pair.Key)
		if count == 10 { // TODO сделать гибче
			break
		}
	}
	fmt.Fprintf(w, "</table>")
}

func main() {
	http.HandleFunc(READ_URL, readSiteHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
