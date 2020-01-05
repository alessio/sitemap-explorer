package main

import (
	"github.com/alessio/sitemap-explorer/downloader"
	"github.com/alessio/sitemap-explorer/utils"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

var (
	workers     int
	timeout     int
	printErrors bool
)

func init() {
	flag.IntVar(&workers, "j", 5, "Limit concurrent requests")
	flag.IntVar(&timeout, "t", 30, "HTTP request timeout")
	flag.BoolVar(&printErrors, "e", false, "Print errors to stderr")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"Usage: %s [OPTION]... DOMAIN URL [URL]...\n"+
				"Scrape URLs.\n\n", os.Args[0])
		flag.PrintDefaults()
	}

}

func printResult(r *downloader.Result) {
	obj, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(obj))
}

func crawl(urls chan string, results chan *downloader.Result, errors chan error,
	allowedDomain string, dder downloader.Downloader, timeout time.Duration) {
	for {
		select {
		case u := <-urls:
			res, err := dder.Download(u)
			if err != nil {
				errors <- err
				continue
			}
			results <- res
			for _, child := range res.Links {
				go func(child string, result string, errors chan error,
					allowedDomain string, dder downloader.Downloader) {
					absoluteURL, err := utils.BuildAbsoluteURL(u, child)
					if err != nil {
						return
					}
					if utils.IsAllowedDomain(allowedDomain, absoluteURL) {
						urls <- absoluteURL.String()
					}
				}(child, res.URL, errors, allowedDomain, dder)
			}
		case <-time.After(time.Second * timeout):
			return
		}
	}
}

func mustValidateArgs() {
	flag.Parse()
	if workers < 1 {
		log.Fatalf("error: at least 1 worker is required: %d", workers)
	}
	if timeout < 0 {
		log.Fatalf("error: timeout must be greater than 0: %d", timeout)
	}
	if flag.NArg() < 2 {
		log.Fatalf("error: invalid number of arguments: %d", flag.NArg())
	}
}

func main() {
	mustValidateArgs()
	results := make(chan *downloader.Result)
	urls := make(chan string)
	errors := make(chan error)
	wg := sync.WaitGroup{}
	allowedDomain := flag.Args()[0]
	startUrls := flag.Args()[1:]
	dder := downloader.NewWebPageDownloader()

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			crawl(urls, results, errors, allowedDomain, dder, time.Duration(timeout))
		}()
	}
	go func() {
		for {
			select {
			case e := <-errors:
				if printErrors {
					log.Println(e)
				}
			case r := <-results:
				printResult(r)
			}
		}
	}()
	for _, u := range startUrls {
		urls <- u
	}
	wg.Wait()
	close(urls)
	close(errors)
	close(results)
}
