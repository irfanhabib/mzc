package fetcher

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

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
	idleMutex     sync.RWMutex
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

	f.idleMutex.RLock()
	isIdle := f.idle
	f.idleMutex.RUnlock()
	return isIdle
}

// InputChannel channel where the worker receives links to crawl
func (f *BasicFetcher) InputChannel() chan string {
	return f.inputChannel
}

func (f *BasicFetcher) setIdle(idle bool) {
	f.idleMutex.Lock()
	f.idle = idle
	f.idleMutex.Unlock()
}

// OutputChannel channel where the worker responds with crawled links
func (f *BasicFetcher) OutputChannel() chan *URLMap {
	return f.outputChannel
}

// Run main worker loop
func (f *BasicFetcher) Run() {

	for {
		url := <-f.inputChannel
		f.setIdle(false)
		links, _ := f.fetch(url)
		f.outputChannel <- &URLMap{URL: url, Links: links}
		f.setIdle(true)

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
	linksMap := make(map[string]struct{})
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
						link, err := f.normalise(a.Val, baseURL)
						if err == nil {
							linksMap[link.String()] = struct{}{}
							break

						}
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
			parsedK, err := url.Parse(k)
			if err == nil {
				uniqueLinks[idx] = parsedK
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

	parsedLink, err := url.Parse(link)
	if err != nil {
		return true
	}
	if parsedLink.Fragment != "" {
		return true
	}
	if parsedLink.IsAbs() {
		if !strings.HasSuffix(parsedLink.Host, f.baseDomain) {
			return true
		}
	}
	return false
}
