// analize-site-to-pdf project main.go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const READ_URL = "/read/"

type Page struct {
	Text  string
	Table map[string]int
}

func (p *Page) toCount() {

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

	fmt.Println(string(body))
	return &Page{Text: string(body)}, nil
}

func readSiteHandler(w http.ResponseWriter, r *http.Request) {
	siteUrl := r.URL.Path[len(READ_URL):]
	p, _ := downloadPage(siteUrl)
	// отправить пдф обратно
	fmt.Fprintf(w, "<h1>Table</h1><div>%s</div>", p.Table)
}

func main() {
	http.HandleFunc(READ_URL, readSiteHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
