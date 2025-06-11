package crawler

import (
	"fmt"
	"io"
	"net/http"
	url2 "net/url"
	"search_egine/db"
	"search_egine/parser"
	"strings"
	"sync"
	"time"
)

const MinTimeBetweenRequest = 40 * time.Second
const MaxIterationForGetUrl = 100000
const MaxSizeQueue = 1000000

var urlChanSender = make(chan string, 1)
var urlChanReceiver = make(chan []string, 1)
var Wg = &sync.WaitGroup{}

type Queue struct {
	Urls    []string
	Visited map[string]interface{}
	Domains map[string]*Domain
	mu      sync.Mutex
	start   time.Time
}
type Domain struct {
	RobotTxt        *parser.RobotTxt
	LastVisitedTime time.Time
}

func NewQueue() *Queue {
	return &Queue{
		Urls:    []string{},
		Visited: make(map[string]interface{}),
		Domains: make(map[string]*Domain),
		mu:      sync.Mutex{},
		start:   time.Now(),
	}
}

func (q *Queue) AddUrl(urls []string) {
	q.mu.Lock()

	defer func() {
		if len(q.Urls) > MaxSizeQueue {
			q.Urls = q.Urls[MaxSizeQueue-10000:]
			q.Visited = make(map[string]interface{})
			q.Domains = make(map[string]*Domain)
		}

		q.mu.Unlock()
	}()
	for _, url := range urls {
		if _, ok := q.Visited[url]; ok {
			continue
		}
		urlParsed, err := url2.Parse(url)
		if err != nil {
			continue
		}
		if _, ok := q.Domains[urlParsed.Host]; !ok {
			q.Domains[urlParsed.Host] = &Domain{
				RobotTxt:        &parser.RobotTxt{},
				LastVisitedTime: time.Now(),
			}
		}

		q.Urls = append(q.Urls, url)
	}

}
func (q *Queue) GetUrl() string {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.Urls) == 0 {
		return ""
	}

	url := ""
	for i := 0; i < MaxIterationForGetUrl && i < len(q.Urls); i++ {
		url = q.Urls[0]
		if _, ok := q.Visited[url]; ok {
			q.Urls = q.Urls[1:]
			continue
		}
		urlParsed, err := url2.Parse(q.Urls[0])
		if err != nil {
			q.Urls = q.Urls[1:]
			continue

		}
		if d, ok := q.Domains[urlParsed.Host]; ok {
			if time.Now().Sub(d.LastVisitedTime) < MinTimeBetweenRequest && time.Now().Sub(q.start) > time.Second*10 {

				q.Urls = q.Urls[1:]
				q.Urls = append(q.Urls, url)

				continue
			}
			if d.RobotTxt.CheckIfIsDisAllowPath(url) {
				q.Urls = q.Urls[1:]
				continue
			}
		}

		// Parce que je supprime les domaines a partir d'un moment ici je m'assure que le domaine existe

		if _, ok := q.Domains[urlParsed.Host]; !ok {
			q.Domains[urlParsed.Host] = &Domain{
				RobotTxt:        &parser.RobotTxt{},
				LastVisitedTime: time.Now(),
			}
			q.Domains[urlParsed.Host].RobotTxt.GetDisallowPath(url)
		}
		q.Urls = q.Urls[1:]

		q.Domains[urlParsed.Host].LastVisitedTime = time.Now()
		q.Visited[url] = nil
		return url
	}
	return url

}
func (q *Queue) QueueHandler() {
	Wg.Add(2)
	go func() {
		defer Wg.Done()
		for {
			url := q.GetUrl()
			urlChanSender <- url
		}
	}()
	go func() {
		defer Wg.Done()
		for {
			url := <-urlChanReceiver

			q.AddUrl(url)
		}

	}()

}

func CrawlerProcess(id int) {
	storage, err := db.NewStorage("search_engine")
	if err != nil {
		return
	}
	Wg.Add(1)
	defer Wg.Done()
	for {
		url := <-urlChanSender
		if url == "" {
			time.Sleep(time.Second * 1)
		}
		fmt.Printf("Crawler %d : %s\n", id, url)
		resp, err := http.Get(url)
		if err != nil {

			continue
		}
		if resp.StatusCode != 200 || !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
			resp.Body.Close()
			continue
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			continue
		}
		dataStr := string(data)
		p := parser.NewParser(dataStr, url)
		p.Traverse()
		urls := make([]string, 0)
		page := db.Page{
			Url:     url,
			Content: p.Content,
			Urls:    p.Url,
		}
		storage.Store(&page)
		for url, _ := range p.Url {
			urls = append(urls, url)
		}
		urlChanReceiver <- urls

	}

}
