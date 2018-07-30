MZC: The Monzo Take Home test Web Crawler

This is a web crawler written for the monzo take home test.

`mzc` is a configurable concurrent crawler that produces a Site Map. The Site Map consists of unique child links from a page. 


**Usage**

To run `mzc`, execute the following:
```
go get github.com/irfanhabib/mzc
```

Command line options include:
```
$ mzc --help
usage: mzc [<flags>] <url>

Flags:
      --help                  Show context-sensitive help (also try --help-long and --help-man).
  -v, --verbose               Verbose mode.
  -w, --workers=50            Number of concurrent workers
  -o, --output="sitemap.txt"  Output file name
  -d, --debug                 Enable CPU profiling
      --ignore-robots-txt     Ignore Robots.txt

Args:
  <url>  URL to crawl.
```

````
$ mzc https://www.monzo.com -o monzo.txt
INFO[0000] Instantiating SiteMap Generator...
INFO[0000] SiteMap Generator running.
INFO[0000] Instantiating Scheduler...
INFO[0000] Scheduler running.
INFO[0000] Start crawling https://www.monzo.com
INFO[0005] Sitemap has been generated. Please see the file monzo.txt
INFO[0005] Time taken: 5.520834449s
````

Site Map format looks like this:
```
$ cat monzo.txt
URL:https://www.monzo.com
	Links:
	URL:https://www.monzo.com/careers
		Links:
		URL:https://www.monzo.com/blog/2017/03/10/transparent-by-default/
			Links:
			URL:https://www.monzo.com/static/css/blog.css?v21
				Links: N/A
			URL:https://www.monzo.com/blog/monzo-hq
				Links:
				URL:https://www.monzo.com/blog/2018/05/22/big-news-about-small-print/
					Links:
					URL:https://www.monzo.com/blog/authors/harry-ashbridge
						Links: N/A
				URL:https://www.monzo.com/blog/2018/06/29/typeform-breach/
					Links: N/A
				URL:https://www.monzo.com/blog/2018/06/28/ticketmaster-breach/
					Links:
					URL:https://www.monzo.com/blog/authors/natasha-vernier
...
```
**Tests**

To run all tests, execute the following from the root:
```
$ go tests ./...
```

**Architecture**

See [Diagram](docs/architecture.png)