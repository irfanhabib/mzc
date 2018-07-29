package sitemap

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/irfanhabib/mzc/fetcher"
)

type SiteMap interface {
	Run()
	InputChannel() chan *fetcher.URLMap
	Print()
}

type MapSiteMapImpl struct {
	inputChannel      chan *fetcher.URLMap
	rootUrl           string
	linksMap          sync.Map
	siteMapFileWriter *bufio.Writer
	outputFileName    string
}

func New(inputChannel chan *fetcher.URLMap, rootUrl string, outputFileName string) SiteMap {
	return &MapSiteMapImpl{inputChannel: inputChannel, rootUrl: rootUrl, outputFileName: outputFileName}
}

func (this *MapSiteMapImpl) Run() {

	f, err := os.OpenFile(this.outputFileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	this.siteMapFileWriter = bufio.NewWriter(f)

	for {
		link := <-this.inputChannel
		this.linksMap.LoadOrStore(link.URL, link)
	}
}

func (this *MapSiteMapImpl) InputChannel() chan *fetcher.URLMap {
	return this.inputChannel
}

func (this *MapSiteMapImpl) Print() {

	node, ok := this.linksMap.Load(this.rootUrl)
	if !ok {
		log.Error("Unable to load root URL. This means scheduler failed to craw")
		os.Exit(-1)
	}

	str := this._Print(node.(*fetcher.URLMap), 0, make(map[string]bool))
	this.siteMapFileWriter.WriteString(str)
	this.siteMapFileWriter.Flush()

}

func (this *MapSiteMapImpl) _Print(node *fetcher.URLMap, level int, seenLinks map[string]bool) string {

	nodeString := fmt.Sprintf("%s\n", node.URL)
	log.Debugf("Processing URL: %s\n", node.URL)

	for _, childUrl := range node.Links {
		baseSep := strings.Repeat("\t", level+1)
		if _, ok := seenLinks[childUrl.String()]; ok {
			// Don't recurse
			continue
		}
		childNode, ok := this.linksMap.Load(childUrl.String())
		if ok {
			// Add all current level children to seen links
			ignoreURLMap := make(map[string]bool)
			for _, childURI := range node.Links {
				ignoreURLMap[childURI.String()] = true
				ignoreURLMap[fmt.Sprintf("%s/", childURI.String())] = true
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
