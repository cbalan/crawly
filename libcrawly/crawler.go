package libcrawly

import (
	"fmt"
	"time"
	"sync"
	"github.com/PuerkitoBio/goquery"
	"log"
)

type Crawler struct{
	wg          sync.WaitGroup
	seenLock    sync.RWMutex
	seenMap map[string]bool
	workChan    chan string
	ValidUrl func (string) bool
}

func NewCrawler(workChanBufferSize int) *Crawler {
	return &Crawler{
		seenMap: make(map[string]bool),
		workChan:  make(chan string, workChanBufferSize),
		ValidUrl: func(url string) bool {
			return true
		},
	}
}

func (c *Crawler) Crawl(url string, workersSize int) {
	c.startWorkers(workersSize)
	c.pushUrl(url)
	c.wait()
}

func (c *Crawler) startWorkers(workersSize int) {
	for i := 0; i < workersSize; i++ {
		c.wg.Add(1)
		go func(i int) {
			defer c.wg.Done()
			c.work(fmt.Sprintf("%d", i), 5*time.Second)
		}(i)
	}
}

func (c *Crawler) wait() {
	c.wg.Wait()
}

func (c *Crawler) pushUrl(url string) {
	if !c.ValidUrl(url) {
		return
	}

	if _, exists := c.seenMap[url]; exists {
		return
	}

	c.seenLock.Lock()
	c.seenMap[url] = true
	c.seenLock.Unlock()

	c.workChan<-url
}

func (c *Crawler) popUrl(timeout time.Duration) (string, bool) {
	url := ""; ok := true
	timeoutChan := time.After(timeout)
	select {
	case url = <-c.workChan :
	case <-timeoutChan:
		ok = false
	}

	return url, ok
}

func (c *Crawler) work(name string, timeout time.Duration) {
	for {
		if url, ok := c.popUrl(timeout); ok {
			timeBefore := time.Now()
			doc, err := goquery.NewDocument(url)
			latency := time.Since(timeBefore)
			if err != nil {
				log.Printf("[%s] %.3fs %s @error:%s\n", name, latency.Seconds(), url, err)
				continue
			}
			log.Printf("%d [%s] %.3fs %s \n", len(c.workChan), name, latency.Seconds(), url)
			go func(doc *goquery.Document) {
				doc.Find("a").Each(func(i int, s *goquery.Selection) {
					if link, exists := s.Attr("href"); exists {
						c.pushUrl(link)
					}
				})
			}(doc)
		} else {
			fmt.Printf("[%s] exit\n", name)
			break
		}
	}
}
