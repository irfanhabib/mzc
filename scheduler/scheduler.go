package scheduler

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/irfanhabib/mzc/fetcher"
	"github.com/irfanhabib/mzc/sitemap"
)

type BasicScheduler struct {
	workersCount  int
	workers       []fetcher.Fetcher
	visitedLinks  sync.Map
	queueChannel  chan string
	mainChannel   chan string
	siteMap       sitemap.SiteMap
	submittedJobs int
	rand          *rand.Rand
	Completed     chan bool
	domain        string
}

func New(mainChannel chan string, siteMap sitemap.SiteMap, crawlUrl string) *BasicScheduler {

	urlStruct, err := url.Parse(crawlUrl)
	if err != nil {
		os.Exit(-1)
	}
	return &BasicScheduler{
		workersCount: 10,
		mainChannel:  mainChannel,
		siteMap:      siteMap,
		Completed:    make(chan bool),
		domain:       urlStruct.Host,
	}
}

func (this *BasicScheduler) init() {
	this.workers = make([]fetcher.Fetcher, 10)
	for i := 0; i < this.workersCount; i++ {
		this.workers[i] = fetcher.New(this.domain, make(chan string, 10), make(chan *fetcher.URLMap, 10))
		go this.workers[i].Run()
	}

	this.rand = rand.New(rand.NewSource(time.Now().UnixNano()))

}

func (this *BasicScheduler) getWorker() fetcher.Fetcher {
	val := int(this.rand.Float64() * 10)
	return this.workers[val]
}
func (this *BasicScheduler) Run() {

	// Setup 10 workers
	this.init()

	// submit units of work
	// go this.submitLinks()

	for i := 0; i < this.workersCount; i++ {
		go this.analyseLinksForWorker(i)
	}

	url := <-this.mainChannel
	this.getWorker().InputChannel() <- url

	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			// monitor worker idleness
			systemIdle := true
			for i := 0; i < 10; i++ {
				systemIdle = systemIdle && this.workers[i].Idle()
				if !systemIdle {
					break
				}
			}
			if systemIdle == true {
				this.Completed <- true
			}
		}
	}()
}

func (this *BasicScheduler) analyseLinksForWorker(index int) {

	for {
		links := <-this.workers[index].OutputChannel()

		// Filter out visited Links
		tmpLinks := links.Links
		newLinks := make([]string, len(tmpLinks))
		idx := 0
		for _, childLink := range tmpLinks {
			urlObj, _ := url.Parse(childLink)
			_, ok := this.visitedLinks.Load(urlObj.String())
			if !ok {
				newLinks[idx] = urlObj.String()
				idx++
			}
		}
		links.Links = newLinks[0:idx]
		this.siteMap.InputChannel() <- links
		go func() {
			for _, urlStr := range links.Links {
				this.visitedLinks.Store(urlStr, true)
				this.getWorker().InputChannel() <- urlStr
				fmt.Printf("Saved URL %s\n", urlStr)
			}
		}()
	}
}
