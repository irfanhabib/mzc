package sitemap

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/irfanhabib/mzc/fetcher"
)

type MapSiteMapImpl struct {
	inputChannel chan *fetcher.URLMap
	rootUrl      string
	linksMap     sync.Map
}

func New(inputChannel chan *fetcher.URLMap, rootUrl string) SiteMap {
	return &MapSiteMapImpl{inputChannel: inputChannel, rootUrl: rootUrl}
}

func (this *MapSiteMapImpl) Run() {
	f, err := os.OpenFile("logFile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	for {
		link := <-this.inputChannel
		this.linksMap.LoadOrStore(link.URL, link)
	}
}

func (this *MapSiteMapImpl) InputChannel() chan *fetcher.URLMap {
	return this.inputChannel
}

func (this *MapSiteMapImpl) Print() string {

	node, ok := this.linksMap.Load(this.rootUrl)
	if !ok {
		fmt.Println("Something went wrong!")
	}

	str := this._Print(node.(*fetcher.URLMap), 0, make(map[string]bool))
	log.Printf("\n%s", str)
	return str

}

func (this *MapSiteMapImpl) _Print(node *fetcher.URLMap, level int, seenLinks map[string]bool) string {

	nodeString := fmt.Sprintf("%s\n", node.URL)
	fmt.Printf("Processing URL: %s\n", node.URL)

	for _, childUrl := range node.Links {
		baseSep := strings.Repeat("\t", level+1)
		if _, ok := seenLinks[childUrl]; ok {
			// Don't recurse
			continue
		}
		childNode, ok := this.linksMap.Load(childUrl)
		if ok {
			// Add all current level children to seen links
			ignoreURLMap := make(map[string]bool)
			for _, childURI := range node.Links {
				ignoreURLMap[getURL(childURI)] = true
				ignoreURLMap[fmt.Sprintf("%s/", getURL(childURI))] = true
			}
			ignoreURLMap[getURL(node.URL)] = true
			ignoreURLMap[fmt.Sprintf("%s/", getURL(node.URL))] = true

			for k, v := range seenLinks {
				ignoreURLMap[k] = v
			}
			nodeString = fmt.Sprintf("%s%s%s\n", nodeString, baseSep, this._Print(childNode.(*fetcher.URLMap), level+1, ignoreURLMap))
		} else {
			// Leaf node
			nodeString = fmt.Sprintf("%s%s%s\n", nodeString, baseSep, childUrl)
		}
	}
	return nodeString
}

func getURL(stringURL string) string {
	URL, _ := url.Parse(stringURL)
	return URL.String()
}
