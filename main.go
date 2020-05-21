// analize-site-to-pdf project main.go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"bytes"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/grokify/html-strip-tags-go" // => strip
	"github.com/signintech/gopdf"
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

	// Создание pdf
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 595.28, H: 841.89}}) //595.28, 841.89 = A4
	pdf.AddPage()
	err := pdf.AddTTFFont("OPEN SANS", "./ttf/OpenSans-Regular.ttf")
	if err != nil {
		log.Print(err.Error())
		return
	}
	err = pdf.SetFont("OPEN SANS", "", 14)
	if err != nil {
		log.Print(err.Error())
		return
	}
	pdf.SetGrayFill(0.5)

	fmt.Fprintf(w, "<h1>Table</h1><table>") // Отображение на странице

	for count, pair := range sortedPairs {
		b := new(bytes.Buffer)
		fmt.Fprintf(b, "%s: %d", pair.Key, pair.Value)
		pdf.Cell(nil, b.String())
		pdf.Br(20)

		fmt.Fprintf(w, "<tr><td>%d</td><td>%s</td></tr>", pair.Value, pair.Key)
		if count == 10 { // TODO сделать гибче
			break
		}
	}
	fmt.Fprintf(w, "</table>")

	pdf.WritePdf("stats.pdf")
}

func main() {
	http.HandleFunc(READ_URL, readSiteHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
