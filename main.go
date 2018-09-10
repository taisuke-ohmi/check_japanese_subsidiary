package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	iconv "github.com/djimenez/iconv-go"
)

const GOOGLEPLAY = `https://play.google.com/store/apps/details`

func main() {
	corp := "Jam City, Inc."
	if checkJapaneseSubsidiary(corp) {
		fmt.Println(corp)
	}
}

func checkJapaneseSubsidiary(corp string) bool {
	str := url.QueryEscape(corp)
	// Request the HTML page.
	res, err := http.Get(fmt.Sprintf("https://www.google.co.jp/search?q=%s", str))
	if err != nil {
		log.Fatalf("failed to http get. %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// convert to utf8
	utfBody, err := iconv.NewReader(res.Body, "shift_jis", "utf-8")
	if err != nil {
		log.Fatalf("failed to decode: %v", err)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(utfBody)
	if err != nil {
		log.Fatalf("failed to load html. %v", err)
	}

	var isJapaneseSite bool
	var isGplaySite bool
	r := regexp.MustCompile(GOOGLEPLAY)
	doc.Find(".g").Each(func(i int, s *goquery.Selection) {
		if i > 2 || isGplaySite || isJapaneseSite {
			return
		}

		// whether or not first result url match https://play.google.com/store/apps/details
		url := s.Find("cite").Text()
		if r.MatchString(url) {
			isGplaySite = true
			return
		}

		// whether or not title is alphabet
		title := s.Find(".r").Find("a").Text()
		if !isAlphabet(title) {
			isJapaneseSite = true
		}
	})

	return isJapaneseSite
}

// return true if arg is alphabet
func isAlphabet(s string) bool {
	if len(s) == utf8.RuneCountInString(s) {
		return true
	}
	return false
}
