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

// SiteMap site map interface
type SiteMap interface {
	Run()
	InputChannel() chan *fetcher.URLMap
	Print()
}

// MapSiteMapImpl basic sitemap implementation that uses a map for storing links data
type MapSiteMapImpl struct {
	inputChannel      chan *fetcher.URLMap
	rootURL           string
	linksMap          sync.Map
	siteMapFileWriter *bufio.Writer
	outputFileName    string
}

// New instantiates a new instance
func New(inputChannel chan *fetcher.URLMap, rootURL string, outputFileName string) SiteMap {
	return &MapSiteMapImpl{inputChannel: inputChannel, rootURL: rootURL, outputFileName: outputFileName}
}

// Run main loop
func (sm *MapSiteMapImpl) Run() {

	f, err := os.OpenFile(sm.outputFileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	sm.siteMapFileWriter = bufio.NewWriter(f)

	for {
		link := <-sm.inputChannel
		sm.linksMap.LoadOrStore(link.URL, link)
	}
}

// InputChannel input channel for sending link data to a SiteMap instance
func (sm *MapSiteMapImpl) InputChannel() chan *fetcher.URLMap {
	return sm.inputChannel
}

// Print prints sitemap data
func (sm *MapSiteMapImpl) Print() {

	node, ok := sm.linksMap.Load(sm.rootURL)
	if !ok {
		log.Error("Unable to load root URL. sm means scheduler failed to craw")
		os.Exit(-1)
	}

	str := sm._Print(node.(*fetcher.URLMap), 0, make(map[string]struct{}))
	sm.siteMapFileWriter.WriteString(str)
	sm.siteMapFileWriter.Flush()

}

func (sm *MapSiteMapImpl) _Print(node *fetcher.URLMap, level int, seenLinks map[string]struct{}) string {

	nodeString := fmt.Sprintf("%s\n", node.URL)
	log.Debugf("Processing URL: %s\n", node.URL)
	baseSep := strings.Repeat("\t", level+1)

	childLinksString := ""
	for _, childURL := range node.Links {
		if _, ok := seenLinks[childURL.String()]; ok {
			// Don't recurse
			continue
		}
		childNode, ok := sm.linksMap.Load(childURL.String())
		if ok {
			// Add all current level children to seen links
			ignoreURLMap := make(map[string]struct{})
			for _, childURI := range node.Links {
				addURLToMap(ignoreURLMap, childURI.String())
			}
			addURLToMap(ignoreURLMap, getURL(node.URL))

			for k, v := range seenLinks {
				ignoreURLMap[k] = v
			}
			childLinksString = fmt.Sprintf("%s%s%s", childLinksString, baseSep, sm._Print(childNode.(*fetcher.URLMap), level+1, ignoreURLMap))
		} else {
			// Leaf node
			childLinksString = fmt.Sprintf("%s%s%s\n", childLinksString, baseSep, childURL)
		}
	}
	if strings.Trim(childLinksString, " \n\t") == "" {
		nodeString = fmt.Sprintf("URL:%s%sLinks: N/A\n", nodeString, baseSep)

	} else {
		nodeString = fmt.Sprintf("URL:%s%sLinks:\n%s", nodeString, baseSep, childLinksString)
	}
	return nodeString
}

func addURLToMap(ignoreURLMap map[string]struct{}, URL string) {
	ignoreURLMap[URL] = struct{}{}
	ignoreURLMap[fmt.Sprintf("%s/", URL)] = struct{}{}
}

func getURL(stringURL string) string {
	URL, _ := url.Parse(stringURL)
	return URL.String()
}
