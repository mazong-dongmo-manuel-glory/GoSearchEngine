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

var visited = make(map[string]bool)

const MaxDomainSize = 1000000
const MaxQueueSize = 10000
const MinSize = 1000
const MaxVisitedSize = 10000

var domains = Domains{}
var dbName = "search_engine"

type Crawler struct{}

func (cr *Crawler) Crawl(id int) {

	rb := &parser.RobotTxt{}
	storage, err := db.NewStorage(dbName)
	if err != nil {
		panic(err)
		return
	}

	for domains.Size() > 0 {
		lock.Lock()
		if domains.Size() == 0 {
			lock.Unlock()
			break
		}

		// Récupère une URL et la retire de la file
		urlToCrawl := GetUrlInQueue()
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
			if !isVisited && rb.CheckIfIsDisAllowPath(url) == false {
				AddNewUrlInQueue(url)

			}
			lock.Unlock()

		}
	}

}
func QeueHandler() {
	storage, err := db.NewStorage(dbName)
	if err != nil {
		panic(err)
		return
	}
	for {
		fmt.Println()
		fmt.Println()

		time.Sleep(20 * time.Second)
		fmt.Printf("Queue size : %v\n", len(queue))
		fmt.Printf("Domains size : %v\n", domains.Size())
		fmt.Println()

		fmt.Println()
		if domains.Size() > MaxDomainSize {
			domains = Domains{}
		}
		if len(visited) > MaxVisitedSize {
			visited = make(map[string]bool)
		}
		if len(queue) > MaxQueueSize+MinSize {
			urlToStore := queue[:MaxQueueSize-MinSize-1]
			storage.StoreQueue(urlToStore)
			queue = queue[MaxQueueSize-MinSize-1:]

		} else if len(queue) > MinSize {
			newUrlToQueue := storage.GetQueue(MinSize)
			queue = append(queue, newUrlToQueue...)
		}

	}
}

func Init(urls []string) {
	for _, url := range urls {
		AddNewUrlInQueue(url)
	}
}
