package main

import (
	"time"

	"github.com/irfanhabib/mzc/fetcher"
	"github.com/irfanhabib/mzc/scheduler"
	"github.com/irfanhabib/mzc/sitemap"
	log "github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	verbose        = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	url            = kingpin.Arg("url", "URL to crawl.").Required().String()
	workers        = kingpin.Flag("workers", "Number of concurrent workers").Short('w').Default("10").Int()
	outputFileName = kingpin.Flag("output", "Output file name").Short('o').Default("sitemap").String()
)

func main() {

	kingpin.Parse()

	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	startTime := time.Now()
	mainChannel := make(chan string)
	// Buffered channel for receiving crawled links from workers
	siteMapChannel := make(chan *fetcher.URLMap, *workers)

	log.Infof("Instantiating SiteMap Generator...")
	siteMap := sitemap.New(siteMapChannel, *url, *outputFileName)
	go siteMap.Run()
	log.Infof("SiteMap Generator running.")

	log.Infof("Instantiating Scheduler...")
	sched := scheduler.New(mainChannel, siteMap, *url, *workers)
	// `done` channel will prevent premature exiting of program
	done := sched.Completed()
	go sched.Run()
	log.Infof("Scheduler running.")

	// Fire off crawling
	log.Infof("Start crawling %s", *url)
	mainChannel <- *url

	// Wait till all workers are idle
	<-done
	siteMap.Print()
	log.Infof("Sitemap has been generated. Please see the file %s", *outputFileName)
	log.Infof("Time taken: %s", time.Since(startTime))
}
