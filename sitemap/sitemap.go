package sitemap

import (
	"fmt"
	"strings"

	"github.com/irfanhabib/mzc/fetcher"
)

type SiteMap interface {
	Run()
	InputChannel() chan *fetcher.URLMap
	Print() string
}

type SiteMapImpl struct {
	inputChannel chan *fetcher.URLMap
	root         *SiteMapTreeNode
	stash        map[string]*fetcher.URLMap
}
type SiteMapTreeNode struct {
	url      string
	children []*SiteMapTreeNode
}

func (this *SiteMapImpl) InputChannel() chan *fetcher.URLMap {
	return this.inputChannel
}

// func New(inputChannel chan *fetcher.URLMap) SiteMap {
// 	return &SiteMapImpl{inputChannel: inputChannel, stash: make(map[string]*fetcher.URLMap)}
// }

func (this *SiteMapImpl) Print() string {

	this.mergeStash()
	return this._Print(this.root, 0, make(map[string]bool))
}

func (this *SiteMapImpl) mergeStash() {

	for link, linkMap := range this.stash {
		node := this.findNode(this.root, link)

		if node == nil {
			fmt.Printf("trouble")
			continue
		}
		children := make([]*SiteMapTreeNode, len(linkMap.Links))
		for i, childUrl := range linkMap.Links {
			children[i] = &SiteMapTreeNode{url: childUrl}
		}
		node.children = children

	}

}

func (this *SiteMapTreeNode) String() string {
	return this.url
}
func (this *SiteMapImpl) _Print(node *SiteMapTreeNode, level int, seenLinks map[string]bool) string {

	str := ""
	if node != nil {

		_, ok := seenLinks[node.url]
		if !ok {
			str = fmt.Sprintf("%s%s\n", str, node.url)
			seenLinks[node.url] = true
			childStr := ""
			for _, k := range node.children {
				tab := strings.Repeat("\t", level)

				subChildStr := this._Print(k, level+1, seenLinks)
				if childStr == "" && subChildStr == "" {
					continue
				}
				childStr = fmt.Sprintf("%s%s%s\n", childStr, tab, subChildStr)
			}
			str = fmt.Sprintf("%s%s", str, childStr)
		}
	}
	return str
}
func (this *SiteMapImpl) Run() {
	for {
		node := this.root

		links := <-this.inputChannel
		if node == nil {
			this.root = &SiteMapTreeNode{url: links.URL}
			node = this.root
		} else {
			node = this.findNode(this.root, links.URL)
		}
		if node == nil {
			// Parent node hasnt been added yet, store in stash
			this.stash[links.URL] = links
			continue
		}
		children := make([]*SiteMapTreeNode, len(links.Links))
		for i, childUrl := range links.Links {
			children[i] = &SiteMapTreeNode{url: childUrl}
		}
		node.children = children

	}
}

func (this *SiteMapImpl) findNode(node *SiteMapTreeNode, url string) *SiteMapTreeNode {

	if node == nil {
		return nil
	}

	if node.url == url {
		return node
	}

	if node.children == nil {
		return nil
	}
	for _, k := range node.children {
		ret := this.findNode(k, url)
		if ret != nil {
			return ret
		}
	}
	return nil
}
