package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

// Crawler crawls all the pages on a specified domain
type Crawler struct {
	complete     sync.WaitGroup
	domain       *url.URL
	urlsToVisit  chan *url.URL
	pagesCrawled chan *Page
}

// NewCrawler creates a new crawler
func NewCrawler(domain *url.URL) *Crawler {
	return &Crawler{
		domain:       domain,
		urlsToVisit:  make(chan *url.URL),
		pagesCrawled: make(chan *Page),
	}
}

// Crawl starts the crawing process and waits for it to complete
func (crawler *Crawler) Crawl() chan *Page {
	results := make(chan *Page)

	go crawler.visitPages()
	go crawler.processPages(results)

	crawler.queueURL(crawler.domain)

	go func() {
		crawler.complete.Wait()
		close(results)
		close(crawler.urlsToVisit)
		close(crawler.pagesCrawled)
	}()

	return results
}

func (crawler *Crawler) queueURL(url *url.URL) {
	crawler.complete.Add(1)
	crawler.urlsToVisit <- url
}

func (crawler *Crawler) visitPages() {
	visitedUrls := map[string]bool{}

	for url := range crawler.urlsToVisit {
		url = crawler.domain.ResolveReference(url)
		if url.Hostname() != crawler.domain.Hostname() {
			crawler.complete.Done()
			continue
		}

		if _, visited := visitedUrls[url.String()]; visited {
			crawler.complete.Done()
			continue
		}

		visitedUrls[url.String()] = true
		go crawler.visitPage(url)
	}
}

func (crawler *Crawler) processPages(results chan<- *Page) {
	for page := range crawler.pagesCrawled {
		if page != nil {
			for _, link := range page.links {
				crawler.queueURL(link)
			}

			results <- page
		}

		crawler.complete.Done()
	}
}

func (crawler *Crawler) visitPage(url *url.URL) {
	resp, err := http.Get(url.String())
	if err != nil {
		crawler.pagesCrawled <- &Page{location: url, err: err}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("non-successful response from '%v', status: '%s'", url, resp.Status)
		crawler.pagesCrawled <- &Page{location: url, err: err}
		return
	}

	if !isHTMLContent(resp) {
		crawler.pagesCrawled <- nil
		return
	}

	crawler.pagesCrawled <- processHTMLContent(url, resp.Body)
}

func isHTMLContent(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	return strings.Contains(contentType, "text/html")
}

func processHTMLContent(url *url.URL, content io.Reader) *Page {
	page := &Page{location: url}

	doc, err := html.Parse(content)
	if err != nil {
		page.err = err
		return page
	}

	visitHTMLNode(doc, func(node *html.Node) {
		page.extractLinksFromNode(node)
	})

	return page
}

func visitHTMLNode(node *html.Node, processHTMLElement func(node *html.Node)) {
	if node.Type == html.ElementNode {
		processHTMLElement(node)
	}

	for childNode := node.FirstChild; childNode != nil; childNode = childNode.NextSibling {
		visitHTMLNode(childNode, processHTMLElement)
	}
}

func (page *Page) extractLinksFromNode(node *html.Node) {
	if node.Data == "a" {
		if href := getAttributeValueAsURL(node, "href"); href != nil {
			page.links = append(page.links, href)
		}
	}
}

func getAttributeValueAsURL(aTag *html.Node, attributeKey string) *url.URL {
	for _, urlAttribute := range aTag.Attr {
		if urlAttribute.Key != attributeKey {
			continue
		}

		rawURL := strings.TrimSpace(urlAttribute.Val)
		if rawURL == "" {
			return nil
		}

		url, err := url.Parse(rawURL)
		if err != nil {
			return nil
		}
		return url
	}
	return nil
}
