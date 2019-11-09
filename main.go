package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Missing Argument: website (e.g. https://bbc.co.uk/)")
	}
	domain, err := url.Parse(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()

	crawler := NewCrawler(domain)
	results := crawler.Crawl()
	pagesCrawled, errorCount := processResults(results)

	elasped := time.Since(start)

	rate := float64(pagesCrawled) / (float64(elasped) / float64(time.Second))
	fmt.Printf("Crawled %d pages in %s (%.2f req/sec) with %d errors \n", pagesCrawled, elasped, rate, errorCount)
}

func processResults(results chan *Page) (pagesCrawled, errorCount int) {
	for page := range results {
		printPage(page)
		if page.err != nil {
			errorCount++
		}
		pagesCrawled++
	}

	return
}

func printPage(page *Page) {
	fmt.Printf("URL:\t%v\n", page.location)

	if page.err != nil {
		fmt.Printf("Error: %v\n", page.err)
		return
	}

	fmt.Println("Links:")
	for _, link := range page.links {
		fmt.Printf("\t%v\n", link)
	}

	fmt.Println()
}
