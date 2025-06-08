package crawler

import (
	"fmt"
	"io"
	"net/http"
	"search_egine/db"
	"search_egine/parser"
	"strings"
	"sync"
	"time"
)

var lock sync.Mutex
var domains = db.Domains{}
var visited = make(map[string]bool)
var dbName = "search_engine"

const MaxSizeVisitedUrl = 10000
const MaxSizeUrlInDomain = 20
const MaxDomain = 100

type Crawler struct{}

func (cr *Crawler) Crawl(id int) {

	rb := &parser.RobotTxt{}
	storage, err := db.NewStorage(dbName)
	if err != nil {
		return
	}
	for domains.Size() > 0 {
		lock.Lock()
		if domains.Size() == 0 {
			lock.Unlock()
			break
		}

		// Récupère une URL et la retire de la file
		urlToCrawl := domains.GetUrlIn()
		visited[urlToCrawl] = true
		lock.Unlock()
		urlToCrawl = strings.TrimSpace(urlToCrawl)

		if urlToCrawl == "" {
			continue
		}

		fmt.Printf("Crawling(%v) : %v\n", id, urlToCrawl)

		rb.GetDisallowPath(urlToCrawl)
		if rb.CheckIfIsDisAllowPath(urlToCrawl) {
			continue
		}

		resp, err := http.Get(urlToCrawl)
		if err != nil {
			//fmt.Printf("Erreur lors de la récupération de l'URL : %v\n", err)
			continue
		}
		if resp.StatusCode != 200 && !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
			//fmt.Printf("Le status n'est pas correct ou le contenu n'est pas du texte : %v\n", resp.StatusCode)
			resp.Body.Close()
			continue
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			//fmt.Printf("Erreur lors de la lecture de l'URL : %v\n", err)
			continue
		}
		dataStr := string(data)

		p := parser.NewParser(dataStr, urlToCrawl)
		p.Traverse()

		page := db.Page{
			Url:     urlToCrawl,
			Content: p.Content,
			Urls:    p.Url,
		}
		storage.Store(&page)
		for url, _ := range p.Url {
			lock.Lock()
			_, isVisited := visited[url]
			if !isVisited && rb.CheckIfIsDisAllowPath(url) == false && len(domains) < MaxDomain {
				domains.AddUrl(url)

			}
			lock.Unlock()

		}
	}

}

func Init(urls []string) {
	for _, url := range urls {
		domains.AddUrl(url)
	}
}

func DomainHandler() {

	storage, err := db.NewStorage(dbName)
	if err != nil {
		return
	}
	for {
		time.Sleep(time.Second * 10)
		lock.Lock()
		if domains.Size() == 0 {
			lock.Unlock()
			break
		}
		if len(visited) > MaxSizeVisitedUrl {
			visited = make(map[string]bool)
		}
		for _, domain := range domains {
			if domain != nil && len(domain.Urls) > MaxSizeUrlInDomain {
				urlToStore := domain.Urls[len(domain.Urls)-(MaxSizeUrlInDomain-4):]
				domain.Urls = domain.Urls[:len(domain.Urls)-(MaxSizeUrlInDomain-4)]
				storage.Store(urlToStore)
			}
		}

		lock.Unlock()

	}
}
