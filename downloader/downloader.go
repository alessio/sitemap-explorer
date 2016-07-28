package downloader

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

// Result defines the attributes of a parsed page
type Result struct {
	URL     string
	Links   []string
	Asssets []string
}

// Downloader defines the interface for a Downloader
type Downloader interface {
	Download(string) (*Result, error)
}

// WebPageDownloader defines the file downloader
type WebPageDownloader struct {
	client       *http.Client
	visited      map[string]bool
	visitedMutex *sync.Mutex
}

// NewWebPageDownloader creates a new WebPageDownloader
func NewWebPageDownloader() *WebPageDownloader {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Transport: tr}
	return &WebPageDownloader{
		client:       &client,
		visited:      make(map[string]bool),
		visitedMutex: &sync.Mutex{},
	}
}

// Download downloads and parse a web page
func (w *WebPageDownloader) Download(u string) (*Result, error) {
	if !w.visitURL(u) {
		return nil, fmt.Errorf("error: %q already visited", u)
	}
	resp, err := w.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	links, assets := w.extractLinks(resp.Body)
	return &Result{URL: u, Links: links, Asssets: assets}, nil
}

// Update visited links
func (w *WebPageDownloader) visitURL(u string) bool {
	w.visitedMutex.Lock()
	defer w.visitedMutex.Unlock()
	if _, ok := w.visited[u]; !ok {
		w.visited[u] = true
		return true
	}
	return false
}

func (w *WebPageDownloader) stripTrailingHash(l string) string {
	var end int

	if !strings.Contains(l, "#") {
		return l
	}
	for i, c := range l {
		if strconv.QuoteRune(c) == "'#'" {
			end = i
			break
		}
	}
	return l[:end]
}

func (w *WebPageDownloader) isStaticAsset(u string) bool {
	ext := filepath.Ext(u)
	if len(ext) != 0 {
		return true
	}
	return false
}

func (w *WebPageDownloader) extractLinks(r io.Reader) ([]string, []string) {
	linksSet := make(map[string]bool)
	assetsSet := make(map[string]bool)
	page := html.NewTokenizer(r)

	for {
		tokenType := page.Next()
		if tokenType == html.ErrorToken {
			links := []string{}
			assets := []string{}
			for k := range linksSet {
				links = append(links, k)
			}
			for k := range assetsSet {
				assets = append(assets, k)
			}
			return links, assets
		}

		token := page.Token()
		if tokenType == html.StartTagToken {
			switch token.DataAtom.String() {
			case "a", "link", "img", "script":
				for _, attr := range token.Attr {
					if attr.Key == "href" || attr.Key == "src" {
						tl := w.stripTrailingHash(attr.Val)
						if strings.Trim(tl, " ") != "" {
							if !w.isStaticAsset(tl) {
								linksSet[tl] = true
							} else {
								assetsSet[tl] = true
							}
						}
					}
				}
			}
		}
	}
}
