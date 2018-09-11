package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

const GOOGLEPLAY = `https://play.google.com/store/apps`

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

	sem := make(chan struct{}, 1)
	logger := log.New(os.Stderr, "ERROR: ", log.LstdFlags)

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		sem <- struct{}{}
		go func(corp string, sem chan struct{}, logger *log.Logger) {
			if checkJapaneseSubsidiary(corp, logger) {
				fmt.Printf("%q,1\n", corp)
			} else {
				fmt.Printf("%q,0\n", corp)
			}
			<-sem
		}(scanner.Text(), sem, logger)
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	for len(sem) > 0 {
		time.Sleep(1000 * time.Millisecond)
	}
}

func checkJapaneseSubsidiary(corp string, logger *log.Logger) bool {
	str := url.QueryEscape(corp)
	// Request the HTML page.
	res, err := http.Get(fmt.Sprintf("https://www.google.co.jp/search?q=%s", str))
	if err != nil {
		logger.Print(fmt.Sprintf("failed to http get. err:%v, corp:%q\n", err, corp))
		return false
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		logger.Print(fmt.Printf("status code error: %d %s corp:%q\n", res.StatusCode, res.Status, corp))
		return false
	}

	utfBody := transform.NewReader(bufio.NewReader(res.Body), japanese.ShiftJIS.NewDecoder())

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(utfBody)
	if err != nil {
		logger.Print(fmt.Sprintf("failed to load html. err:%v, corp:%q\n", err, corp))
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
