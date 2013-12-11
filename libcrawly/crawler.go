package libcrawly

import (
	"time"
	"sync"
	"github.com/PuerkitoBio/goquery"
	"log"
	"errors"
)

type Crawler struct{
	wg          sync.WaitGroup
	seenLock    sync.RWMutex
	seenMap map[string]bool
	workChan    chan string
	ValidUrl func (string) bool
	timeout     time.Duration
}

func NewCrawler(workChanBufferSize int, timeout time.Duration) *Crawler {
	return &Crawler{
		seenMap: make(map[string]bool),
		workChan:  make(chan string, workChanBufferSize),
		ValidUrl: func(string) bool {
			return true
		},
		timeout: timeout,
	}
}

func (c *Crawler) Crawl(url string, workersSize int) {
	c.startWorkers(workersSize)
	c.pushUrl(url, c.timeout)
	c.wait()
}

func (c *Crawler) startWorkers(workersSize int) {
	for i := 0; i < workersSize; i++ {
		c.wg.Add(1)
		go func(i int) {
			defer c.wg.Done()
			c.work(i, c.timeout)
		}(i)
	}
}

func (c *Crawler) wait() {
	c.wg.Wait()
}

func (c *Crawler) pushUrl(url string, timeout time.Duration) (err error) {
	if !c.ValidUrl(url) {
		return
	}

	if _, exists := c.seenMap[url]; exists {
		return
	}

	c.seenLock.Lock()
	c.seenMap[url] = true
	c.seenLock.Unlock()

	select {
	case c.workChan<-url:
	case <-time.After(timeout):
		err = errors.New("pushUrl timeout succedded")
	}

	return err
}

func (c *Crawler) popUrl(timeout time.Duration) (url string, err error) {
	select {
	case url = <-c.workChan :
	case <-time.After(timeout):
		err = errors.New("popUrl timeout succedded")
	}
	return url, err
}

func (c *Crawler) work(workerId int, timeout time.Duration) {
	for {
		if url, err := c.popUrl(timeout); err == nil {
			timeBefore := time.Now()
			doc, err := goquery.NewDocument(url)
			latency := time.Since(timeBefore)
			if err != nil {
				log.Printf("[%d] %.3fs %s @error:%s\n", workerId, latency.Seconds(), url, err)
				continue
			}
			log.Printf("%d [%d] %.3fs %s \n", len(c.workChan), workerId, latency.Seconds(), url)
			go func(doc *goquery.Document) {
				doc.Find("a").Each(func(i int, s *goquery.Selection) {
					if link, exists := s.Attr("href"); exists {
						if err := c.pushUrl(link, timeout); err != nil {
							log.Printf("[%d] pushUrl(%s) failed. %s\n", workerId, link, err)
						}
					}
				})
			}(doc)
		} else {
			log.Printf("[%d] work done. %s\n", workerId, err)
			break
		}
	}
}
