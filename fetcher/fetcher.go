package fetcher

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type Fetcher interface {
	Run()
	InputChannel() chan string
	OutputChannel() chan *URLMap
	Idle() bool
}

type URLMap struct {
	URL   string
	Links []string
}

type BasicFetcher struct {

	// Base URL to detect external links
	baseDomain    string
	inputChannel  chan string
	outputChannel chan *URLMap
	idle          bool
}

func New(baseDomain string, inputChannel chan string, outputChannel chan *URLMap) Fetcher {
	return &BasicFetcher{
		baseDomain:    baseDomain,
		inputChannel:  inputChannel,
		outputChannel: outputChannel,
		idle:          true,
	}
}

func (this *BasicFetcher) Idle() bool {
	return this.idle
}
func (this *BasicFetcher) InputChannel() chan string {
	return this.inputChannel
}

func (this *BasicFetcher) OutputChannel() chan *URLMap {
	return this.outputChannel
}

func (this *BasicFetcher) Run() {

	for {
		url := <-this.inputChannel
		this.idle = false
		links, _ := this.fetch(url)
		this.outputChannel <- &URLMap{URL: url, Links: links}
		this.idle = true
	}
}

func (this *BasicFetcher) fetch(url string) ([]string, error) {

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Exception fetching url: %s", err)
	} else {
		defer response.Body.Close()
		links, err := this.findLinks(response, url)
		if err != nil {
			return nil, err
		}
		return links, nil
	}
}

func (this *BasicFetcher) findLinks(response *http.Response, baseUrl string) ([]string, error) {
	tokenizedHtmlPage := html.NewTokenizer(response.Body)

	// Store links in map to squash dups
	linksMap := make(map[string]bool)
	for {
		token := tokenizedHtmlPage.Next()

		if token == html.ErrorToken {
			break
		}

		if token == html.StartTagToken {
			tag := tokenizedHtmlPage.Token()
			isAnchor := tag.Data == "a"
			if isAnchor {
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
	uniqueLinks := make([]string, len(linksMap))
	idx := 0
	for k := range linksMap {

		if !this.isExternalLink(k, baseUrl) {
			normalisedLink, err := this.normalise(k, baseUrl)
			if err == nil {
				uniqueLinks[idx] = normalisedLink
			}
			idx++
		}
	}
	return uniqueLinks[0:idx], nil

}

func (this *BasicFetcher) normalise(link string, base string) (string, error) {
	linkUrl, err := url.Parse(link)
	if err != nil {
		return "", fmt.Errorf("Unable to normalise %s", link)
	}

	if linkUrl.Fragment != "" {
		return linkUrl.String(), nil
	}

	baseUrl, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("Unable to normalise due to invalid base: %s", baseUrl)
	}

	return baseUrl.ResolveReference(linkUrl).String(), nil
}

func (this *BasicFetcher) isExternalLink(link string, base string) bool {

	// If link is an internal link, exclude from crawling

	normalisedLink, err := this.normalise(link, base)
	if err != nil {
		return true
	}
	linkUrl, err := url.Parse(normalisedLink)
	if err != nil {
		return true
	}

	if linkUrl.Fragment != "" {
		return true
	}
	if linkUrl.IsAbs() {
		if !strings.HasSuffix(linkUrl.Host, this.baseDomain) {
			return true
		}
	}
	return false
}
