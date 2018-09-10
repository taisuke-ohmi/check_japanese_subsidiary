package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	iconv "github.com/djimenez/iconv-go"
)

const GOOGLEPLAY = `https://play.google.com/store/apps/details`

func main() {
	var fp *os.File
	var err error
	if len(os.Args) < 2 {
		log.Fatal("filename is require")
	} else {
		fp, err = os.Open(os.Args[1])
		if err != nil {
			panic(err)
		}
		defer fp.Close()
	}

	sem := make(chan struct{}, 2)

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		sem <- struct{}{}
		go func(corp string, sem chan struct{}) {
			if checkJapaneseSubsidiary(corp) {
				fmt.Printf("%q,1\n", corp)
			} else {
				fmt.Printf("%q,0\n", corp)
			}
			<-sem
		}(scanner.Text(), sem)
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	for len(sem) > 0 {
	}
}

func checkJapaneseSubsidiary(corp string) bool {
	str := url.QueryEscape(corp)
	// Request the HTML page.
	res, err := http.Get(fmt.Sprintf("https://www.google.co.jp/search?q=%s", str))
	if err != nil {
		log.Printf("failed to http get. err:%v, corp:%q\n", err, corp)
		return false
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Printf("status code error: %d %s corp:%q\n", res.StatusCode, res.Status, corp)
		return false
	}

	// convert to utf8
	utfBody, err := iconv.NewReader(res.Body, "shift_jis", "utf-8")
	if err != nil {
		log.Printf("failed to decode: err:%v, corp:%q", err, corp)
		return false
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(utfBody)
	if err != nil {
		log.Printf("failed to load html. err:%v, corp:%q", err, corp)
		return false
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
