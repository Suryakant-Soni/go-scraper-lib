package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func getHref(t html.Token) (ok bool, url string) {
	for _, v := range t.Attr {
		if v.Key == "href" {
			url = v.Val
			ok = true
		}
	}
	// log.Printf("href %v , %v", url, ok)
	return
}

func scrape(url string, chPassUrls chan string, chDone chan bool) {
	defer func() {
		chDone <- true
	}()
	res, err := http.Get(url)
	if err != nil {
		log.Printf("Error %v", err)
		return
	}
	// b, err := io.ReadAll(res.Body)
	// if err != nil {
	// 	log.Println("Error", err)
	// }

	// log.Printf("res.Body %v", string(b))
	// so that the connection can be used by transort for further keep alive connections
	defer res.Body.Close()
	z := html.NewTokenizer(res.Body)
	for {
		tagP := z.Next()
		switch {
		case tagP == html.ErrorToken:
			return
		case tagP == html.StartTagToken:
			// log.Println("getting into start token")
			tag := z.Token()
			isHref := tag.Data == "a"
			// log.Printf("data %v", tag.Data)
			if !isHref {
				continue
			}
			ok, urlHref := getHref(tag)
			if !ok {
				continue
			}
			// log.Printf("index %v", strings.Index(url, "http"))
			isHttp := strings.Index(urlHref, "http") == 0
			if isHttp {
				chPassUrls <- urlHref
			}
		}
	}
}

func main() {
	extractedUrls := make(map[string]bool)
	inputUrls := os.Args[1:]
	log.Printf("urls %v", inputUrls)
	chPassUrls := make(chan string)
	chDone := make(chan bool)
	for _, url := range inputUrls {
		go scrape(url, chPassUrls, chDone)
	}
	for i := 0; i < len(inputUrls); {
		select {
		case url := <-chPassUrls:
			extractedUrls[url] = true
		case <-chDone:
			i++
		}
	}
	log.Printf("total urls found %v \n", len(extractedUrls))
	for url, _ := range extractedUrls {
		log.Printf(":%v", url)
	}
}
