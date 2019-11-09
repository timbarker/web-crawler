package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptyPage(t *testing.T) {

	ts := createTestServerWithSuccessResponses(map[string]string{
		"/": `<html>
				<body>
				</body>
			  </html>`,
	})
	defer ts.Close()

	result := createAndRunCrawler(ts.URL)

	assert.ElementsMatch(t, []*Page{
		&Page{location: URL(ts.URL)},
	}, result)
}

func TestHttpError(t *testing.T) {
	result := createAndRunCrawler("http://localhost:1234")

	require.Equal(t, 1, len(result))
	assert.Equal(t, URL("http://localhost:1234"), result[0].location)
	assert.Error(t, result[0].err)
}

func TestPageWithLinksToOtherDomain(t *testing.T) {

	ts := createTestServerWithSuccessResponses(map[string]string{
		"/": `<html>
				<body>
					<a href="https://other.domain">Link</a>
				</body>
			  </html>`,
	})
	defer ts.Close()

	result := createAndRunCrawler(ts.URL)

	assert.ElementsMatch(t, []*Page{
		&Page{location: URL(ts.URL), links: []*url.URL{URL("https://other.domain")}},
	}, result)
}

func TestPageWithLinksToSameDomain(t *testing.T) {
	ts := createTestServerWithSuccessResponses(map[string]string{
		"/": `<html>
				<body>
					<a href="/subpage">Link</a>
				</body>
			  </html>`,
		"/subpage": `<html>
						<body>
						</body>
					</html>`,
	})
	defer ts.Close()

	result := createAndRunCrawler(ts.URL)

	assert.ElementsMatch(t, []*Page{
		&Page{location: URL(ts.URL), links: []*url.URL{URL("/subpage")}},
		&Page{location: appendPathToURL(URL(ts.URL), "/subpage")},
	}, result)
}

func TestPageWithCircularLinksToSameDomain(t *testing.T) {
	ts := createTestServerWithSuccessResponses(map[string]string{
		"/": `<html>
				<body>
					<a href="/subpage1">Link</a>
				</body>
			  </html>`,
		"/subpage1": `<html>
						<body>
							<a href="/subpage2">Link</a>
						</body>
					 </html>`,
		"/subpage2": `<html>
						<body>
							<a href="/subpage1">Link</a>
						</body>
					 </html>`,
	})
	defer ts.Close()

	result := createAndRunCrawler(ts.URL)

	assert.ElementsMatch(t, []*Page{
		&Page{location: URL(ts.URL), links: []*url.URL{URL("/subpage1")}},
		&Page{location: appendPathToURL(URL(ts.URL), "/subpage1"), links: []*url.URL{URL("/subpage2")}},
		&Page{location: appendPathToURL(URL(ts.URL), "/subpage2"), links: []*url.URL{URL("/subpage1")}},
	}, result)
}

func TestPageWithLinksWithAndWithoutTrailingSlash(t *testing.T) {
	ts := createTestServerWithSuccessResponses(map[string]string{
		"/": `<html>
				<body>
					<a href="/subpage">Link</a>
					<a href="/subpage/">Link</a>
				</body>
			  </html>`,
		"/subpage": `<html>
						<body>
						</body>
					</html>`,
		"/subpage/": `<html>
						<body>
						</body>
					  </html>`,
	})
	defer ts.Close()

	result := createAndRunCrawler(ts.URL)

	assert.ElementsMatch(t, []*Page{
		&Page{location: URL(ts.URL), links: []*url.URL{URL("/subpage"), URL("/subpage/")}},
		&Page{location: appendPathToURL(URL(ts.URL), "/subpage")},
		&Page{location: appendPathToURL(URL(ts.URL), "/subpage/")},
	}, result)
}

func TestNotFoundPage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	result := createAndRunCrawler(ts.URL)

	assert.ElementsMatch(t, []*Page{
		&Page{location: URL(ts.URL), err: fmt.Errorf("non-successful response from '%s', status: '404 Not Found'", ts.URL)},
	}, result)
}

func TestInternalServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	result := createAndRunCrawler(ts.URL)

	assert.ElementsMatch(t, []*Page{
		&Page{location: URL(ts.URL), err: fmt.Errorf("non-successful response from '%s', status: '500 Internal Server Error'", ts.URL)},
	}, result)
}

func TestPageWithNonHTMLContent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
	}))
	defer ts.Close()

	result := createAndRunCrawler(ts.URL)

	assert.ElementsMatch(t, []*Page{}, result)
}

func TestPageWithAnchorTagWithMissingHREF(t *testing.T) {
	ts := createTestServerWithSuccessResponses(map[string]string{
		"/": `<html>
				<body>
					<a target="_blank">Link</a>
				</body>
			 </html>`,
	})
	defer ts.Close()

	result := createAndRunCrawler(ts.URL)

	assert.ElementsMatch(t, []*Page{
		&Page{location: URL(ts.URL)},
	}, result)
}

func TestPageWithAnchorTagWithInvalidURLinHREF(t *testing.T) {
	ts := createTestServerWithSuccessResponses(map[string]string{
		"/": `<html>
				<body>
					<a href=":invalid">Link</a>
				</body>
			 </html>`,
	})
	defer ts.Close()

	result := createAndRunCrawler(ts.URL)

	assert.ElementsMatch(t, []*Page{
		&Page{location: URL(ts.URL)},
	}, result)
}

func TestPageWithAnchorTagWithExtraSpacesInHREF(t *testing.T) {
	ts := createTestServerWithSuccessResponses(map[string]string{
		"/": `<html>
				<body>
					<a href="  /somelink  ">Link</a>
				</body>
			</html>`,
		"/somelink": `<html>
						<body>
						</body>
					 </html>`,
	})
	defer ts.Close()

	result := createAndRunCrawler(ts.URL)

	assert.ElementsMatch(t, []*Page{
		&Page{location: URL(ts.URL), links: []*url.URL{URL("/somelink")}},
		&Page{location: appendPathToURL(URL(ts.URL), "/somelink")},
	}, result)
}

func TestPageWithAnchorTagWithEmptyHREF(t *testing.T) {
	ts := createTestServerWithSuccessResponses(map[string]string{
		"/": `<html>
				<body>
					<a href="">Link</a>
				</body>
			</html>`,
	})
	defer ts.Close()

	result := createAndRunCrawler(ts.URL)

	assert.ElementsMatch(t, []*Page{
		&Page{location: URL(ts.URL)},
	}, result)
}

func createTestServerWithSuccessResponses(responses map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html, ok := responses[r.URL.Path]
		if ok {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintln(w, html)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func createAndRunCrawler(domain string) []*Page {
	var testDomain = URL(domain)
	crawler := NewCrawler(testDomain)

	pagesCrawled := []*Page{}
	for page := range crawler.Crawl() {
		pagesCrawled = append(pagesCrawled, page)
	}
	return pagesCrawled
}

func URL(rawURL string) *url.URL {
	url, _ := url.Parse(rawURL)
	return url
}

func appendPathToURL(url *url.URL, path string) *url.URL {
	return url.ResolveReference(URL(path))
}

func BenchmarkSmallDomain(b *testing.B) {
	ts := createTestServerWithSuccessResponses(createBenchmarkSite(10))
	defer ts.Close()

	for n := 0; n < b.N; n++ {
		createAndRunCrawler(ts.URL)
	}
}

func BenchmarkMediumDomain(b *testing.B) {
	ts := createTestServerWithSuccessResponses(createBenchmarkSite(1000))
	defer ts.Close()

	for n := 0; n < b.N; n++ {
		createAndRunCrawler(ts.URL)
	}
}

func BenchmarkLargeDomain(b *testing.B) {
	ts := createTestServerWithSuccessResponses(createBenchmarkSite(100000))
	defer ts.Close()

	for n := 0; n < b.N; n++ {
		createAndRunCrawler(ts.URL)
	}
}

func createBenchmarkSite(pageCount int) map[string]string {
	page := `<html>
				<body>
					<a href="/%d">Link</a>
				</body>
			</html>`

	site := map[string]string{
		"/": fmt.Sprintf(page, 1),
	}

	for i := 1; i < pageCount; i++ {
		site["/"+strconv.Itoa(i)] = fmt.Sprintf(page, i+1)
	}

	return site
}
