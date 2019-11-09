package main

import (
	"fmt"
	"net/url"
)

// Page stores all the interesting aspects of a crawled web page, such as links
type Page struct {
	location *url.URL
	links    []*url.URL
	err      error
}

func (page *Page) String() string {
	return fmt.Sprintf("Page{location:%v, links: %v, err: %v}", page.location, page.links, page.err)
}
