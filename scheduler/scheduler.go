package scheduler

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/irfanhabib/mzc/fetcher"
	"github.com/irfanhabib/mzc/robottxt"
	"github.com/irfanhabib/mzc/sitemap"
)

// Scheduler interface for Scheduler
type Scheduler interface {
	Run()
	Completed() chan bool
}

// BasicScheduler  basic implementation of a scheduler that orchestrates crawling
// It will randomly send links to crawl to a finite set of workers
type BasicScheduler struct {
	workersCount    int
	workers         []fetcher.Fetcher
	visitedLinks    sync.Map
	queueChannel    chan string
	mainChannel     chan string
	siteMap         sitemap.SiteMap
	submittedJobs   int
	rand            *rand.Rand
	completed       chan bool
	domain          string
	scheme          string
	ignoreRobotsTxt bool
}

// New instiates a new Scheduler instance
func New(mainChannel chan string, siteMap sitemap.SiteMap, crawlURL string, workerCount int, ignoreRobotsTxt bool) Scheduler {

	urlStruct, err := url.Parse(crawlURL)
	if err != nil {
		log.Errorf("Please specify a valid URL (i.e. https://google.com. Error was %+v", err)
		os.Exit(-1)
	}
	return &BasicScheduler{
		workersCount:    workerCount,
		mainChannel:     mainChannel,
		siteMap:         siteMap,
		completed:       make(chan bool),
		domain:          urlStruct.Host,
		scheme:          urlStruct.Scheme,
		ignoreRobotsTxt: ignoreRobotsTxt,
	}
}

// Completed returns the completed channel to
//indicate that the scheduler is done crawling
func (sched *BasicScheduler) Completed() chan bool {
	return sched.completed
}
func (sched *BasicScheduler) init() {

	var robotsTxt robottxt.RobotsTxt
	if !sched.ignoreRobotsTxt {
		robotsTxt = robottxt.New(fmt.Sprintf("%s://%s", sched.scheme, sched.domain))
	}

	sched.workers = make([]fetcher.Fetcher, sched.workersCount)
	for i := 0; i < sched.workersCount; i++ {
		sched.workers[i] = fetcher.New(sched.domain, make(chan string, 10), make(chan *fetcher.URLMap, 10), robotsTxt)
		go sched.workers[i].Run()
	}

	sched.rand = rand.New(rand.NewSource(time.Now().UnixNano()))

}

func (sched *BasicScheduler) getWorker() fetcher.Fetcher {
	val := int(sched.rand.Float64() * float64(sched.workersCount))
	return sched.workers[val]
}

// Run main method that initialises workers and starts the crawling process
func (sched *BasicScheduler) Run() {

	// Setup workers
	sched.init()

	// Instantiate go routine for each worker to
	// receive the crawled link and resubmit child links to other workers
	for i := 0; i < sched.workersCount; i++ {
		go sched.workerHandler(i)
	}

	url := <-sched.mainChannel
	sched.getWorker().InputChannel() <- url

	// gorouting to monitor worker idleness,
	// once all workers are idle no more crawling is taking place
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			systemIdle := true
			for i := 0; i < sched.workersCount; i++ {
				systemIdle = systemIdle && sched.workers[i].Idle()
				if !systemIdle {
					break
				}
			}
			if systemIdle == true {
				sched.completed <- true
			}
		}
	}()
}

func (sched *BasicScheduler) workerHandler(index int) {

	for {
		links := <-sched.workers[index].OutputChannel()

		// Filter out visited Links
		tmpLinks := links.Links
		newLinks := make([]*url.URL, len(tmpLinks))
		idx := 0
		for _, childLink := range tmpLinks {
			_, ok := sched.visitedLinks.Load(childLink.String())
			if !ok {
				newLinks[idx] = childLink
				idx++
			}
		}
		links.Links = newLinks[0:idx]

		sched.siteMap.InputChannel() <- links
		go func() {
			for _, urlStr := range links.Links {
				sched.visitedLinks.Store(urlStr.String(), true)
				sched.getWorker().InputChannel() <- urlStr.String()
				log.Debugf("Crawled URL %s\n", urlStr.String())
			}
		}()
	}
}
