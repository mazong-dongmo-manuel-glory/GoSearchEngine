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

const MaxDomainSize = 1000
const MaxQueueSize = 40000
const MinSize = 10000
const MaxVisitedSize = 1000

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

	for len(queue) > 0 {

		// Récupère une URL et la retire de la file
		lock.Lock()
		urlToCrawl := GetUrlInQueue()
		visited[urlToCrawl] = true
		lock.Unlock()
		urlToCrawl = strings.TrimSpace(urlToCrawl)

		if urlToCrawl == "" {
			fmt.Println("Queue is empty")
			time.Sleep(1 * time.Second)
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
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()

	fmt.Printf("Crawler(%v) is done\n", id)
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()

}
func QeueHandler() {
	storage, err := db.NewStorage(dbName)
	if err != nil {
		panic(err)
		return
	}
	for {

		time.Sleep(5 * time.Second)
		fmt.Println()
		fmt.Println()
		fmt.Printf("Queue size : %v\n", len(queue))
		fmt.Printf("Domains size : %v\n", domains.Size())
		fmt.Printf("Visited size : %v\n", len(visited))
		fmt.Println()
		fmt.Println()
		lock.Lock()
		if len(domains) > MaxDomainSize {
			domains = Domains{}
		}
		if len(visited) > MaxVisitedSize {
			visited = make(map[string]bool)
		}
		if len(queue) > MaxQueueSize {
			urlToStore := queue[len(queue)-(MinSize*2):]
			storage.StoreQueue(urlToStore)
			queue = queue[:len(queue)-(MinSize*2)]

		} else if len(queue) <= MinSize {
			newUrlToQueue := storage.GetQueue(MinSize)
			queue = append(queue, newUrlToQueue...)
		}
		lock.Unlock()

	}
}

func Init(urls []string) {
	for _, url := range urls {
		AddNewUrlInQueue(url)
	}
}
