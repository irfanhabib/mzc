package fetcher

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// Fetcher interface that should be implemented by a worker
type Fetcher interface {
	Run()
	InputChannel() chan string
	OutputChannel() chan *URLMap
	Idle() bool
}

// URLMap struct used by workers to store a URL and its child links
type URLMap struct {
	URL   string
	Links []*url.URL
}

// BasicFetcher basic implementation of a worker
type BasicFetcher struct {
	baseDomain    string
	inputChannel  chan string
	outputChannel chan *URLMap
	idle          bool
}

// New instantiates a worker
func New(baseDomain string, inputChannel chan string, outputChannel chan *URLMap) Fetcher {
	return &BasicFetcher{
		baseDomain:    baseDomain,
		inputChannel:  inputChannel,
		outputChannel: outputChannel,
		idle:          true,
	}
}

// Idle returns the current `idle` value
// A worker is idle if its waiting for work
func (f *BasicFetcher) Idle() bool {
	return f.idle
}

// InputChannel channel where the worker receives links to crawl
func (f *BasicFetcher) InputChannel() chan string {
	return f.inputChannel
}

// OutputChannel channel where the worker responds with crawled links
func (f *BasicFetcher) OutputChannel() chan *URLMap {
	return f.outputChannel
}

// Run main worker loop
func (f *BasicFetcher) Run() {

	for {
		url := <-f.inputChannel
		f.idle = false
		links, _ := f.fetch(url)
		f.outputChannel <- &URLMap{URL: url, Links: links}
		f.idle = true
	}
}

func (f *BasicFetcher) fetch(url string) ([]*url.URL, error) {

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Exception fetching url: %s", err)
	} else {
		defer response.Body.Close()
		links, err := f.findLinks(response, url)
		if err != nil {
			return nil, err
		}
		return links, nil
	}
}

func (f *BasicFetcher) findLinks(response *http.Response, baseURL string) ([]*url.URL, error) {
	tokenizedHTMLPage := html.NewTokenizer(response.Body)

	// Store links in map to squash dups
	linksMap := make(map[string]bool)
	for {
		token := tokenizedHTMLPage.Next()

		if token == html.ErrorToken {
			break
		}

		if token == html.StartTagToken {
			tag := tokenizedHTMLPage.Token()
			isAnchor := tag.Data == "a"
			isLink := tag.Data == "link"
			if isAnchor || isLink {
				for _, a := range tag.Attr {
					if a.Key == "href" {
						linksMap[a.Val] = true
						break
					}
				}
			}
		}
	}

	//extract keys
	uniqueLinks := make([]*url.URL, len(linksMap))
	idx := 0
	for k := range linksMap {

		if !f.isExternalOrInvalidLink(k, baseURL) {
			normalisedLink, err := f.normalise(k, baseURL)
			if err == nil {
				uniqueLinks[idx] = normalisedLink
				idx++
			}
		}
	}
	return uniqueLinks[0:idx], nil

}

func (f *BasicFetcher) normalise(link string, base string) (*url.URL, error) {
	linkURL, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("Unable to normalise %s", link)
	}

	if linkURL.Fragment != "" {
		return linkURL, nil
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("Unable to normalise due to invalid base: %s", baseURL)
	}

	return baseURL.ResolveReference(linkURL), nil
}

func (f *BasicFetcher) isExternalOrInvalidLink(link string, base string) bool {

	// If link is an internal link, exclude from crawling
	normalisedLink, err := f.normalise(link, base)
	if err != nil {
		return true
	}

	if normalisedLink.Fragment != "" {
		return true
	}
	if normalisedLink.IsAbs() {
		if !strings.HasSuffix(normalisedLink.Host, f.baseDomain) {
			return true
		}
	}
	return false
}
