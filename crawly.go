package main

import (
	"strings"
	"flag"
	"runtime"
	"github.com/cbalan/crawly/libcrawly"
)

var (
	baseUrl     = flag.String("url", "http://www.google.com", "Base URL")
	concurrency = flag.Int("c", 1, "Concurrency")
	bufferSize  = flag.Int("bs", 0, "Buffer size")
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	flag.Parse()
	c := libcrawly.NewCrawler(*bufferSize)

	c.ValidUrl = func(url string) bool {
		if !strings.HasPrefix(url, *baseUrl) {
			return false
		}

		for _, value := range []string{"socialshare", "product_compare", "product_compare", "cart", "wishlist"} {
			if strings.Contains(url, value) {
				return false
			}
		}

		return true
	}

	c.Crawl(*baseUrl, *concurrency)
}
