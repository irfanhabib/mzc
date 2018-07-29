package main

import (
	"fmt"

	"github.com/irfanhabib/mzc/fetcher"
	"github.com/irfanhabib/mzc/scheduler"
	"github.com/irfanhabib/mzc/sitemap"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	verbose = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	url     = kingpin.Arg("url", "URL to crawl.").Required().String()
	workers = kingpin.Flag("workers", "Number of concurrent workers").Short('w').Default("10").Int()
)

func main() {

	kingpin.Parse()

	mainChannel := make(chan string, *workers)
	siteMapChannel := make(chan *fetcher.URLMap, *workers)
	siteMap := sitemap.New(siteMapChannel, *url)
	go siteMap.Run()
	sched := scheduler.New(mainChannel, siteMap, *url)
	done := sched.Completed
	go sched.Run()
	mainChannel <- *url
	<-done
	fmt.Printf(siteMap.Print())
}
