package sitemap

import (
	"fmt"
	"testing"

	"github.com/irfanhabib/mzc/fetcher"
)

func TestPrint(t *testing.T) {

	channel := make(chan *fetcher.URLMap)
	siteMap := New(channel, "https://monzo.com")
	go siteMap.Run()

	channel <- &fetcher.URLMap{
		URL:   "https://monzo.com",
		Links: []string{"https://monzo.com/link1", "https://monzo.com/link2"},
	}
	channel <- &fetcher.URLMap{
		URL:   "https://monzo.com/link1",
		Links: []string{"https://monzo.com", "https://monzo.com/link2", "https://monzo.com/link3"},
	}
	p := siteMap.Print()
	fmt.Printf("%s", p)
	<-channel
}
