package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fatalChan := make(chan error)
	r := httprouter.New()
	r.GET("/api/findlinks", FindLinks)

	go func() {
		if err := http.ListenAndServe(":9010", r); err != nil {
			fatalChan <- err
		}
	}()

	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	select {
	case <-term:
		log.Println("Signal terminate detected")
	case err := <-fatalChan:
		log.Fatal("Application failed to run because ", err.Error())
	}
}

func FindLinks(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	url := r.FormValue("url")

	fmt.Fprintf(w, "Page = %q\n", url)

	if len(url) == 0 {
		return
	}

	page, err := parse("https://" + url)
	if err != nil {
		fmt.Printf("error getting page %s: %s\n", url, err)
		return
	}

	links := pageLinks(nil, page)
	for _, link := range links {
		fmt.Fprintf(w, "Link = %q\n", link)
	}

	return
}

func parse(url string) (*html.Node, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("cannot get page")
	}
	b, err := html.Parse(r.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot parse page")
	}
	return b, err
}

func pageLinks(links []string, n *html.Node) []string {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			if a.Key == "href" {
				links = append(links, a.Val)
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		links = pageLinks(links, c)
	}
	return links
}
