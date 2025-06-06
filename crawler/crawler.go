package crawler

import (
	"fmt"
	"io"
	"net/http"
	"search_egine/db"
	"search_egine/parser"
	"strings"
)

type Crawler struct{}

func (cr *Crawler) Crawl(urlToCrawl string) {
	queue := []string{urlToCrawl}
	visited := make(map[string]bool)
	visited[urlToCrawl] = true

	for len(queue) > 0 {
		urlToCrawl := queue[0]
		visited[urlToCrawl] = true
		fmt.Printf("Crawling : %v\n", urlToCrawl)

		queue = queue[1:]

		rb := &parser.RobotTxt{}

		rb.GetDisallowPath(urlToCrawl)

		resp, err := http.Get(urlToCrawl)
		if err != nil {
			fmt.Printf("Erreur lors de la récupération de l'URL : %v\n", err)

			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 && !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
			fmt.Printf("Le status n'est pas correct ou le contenu n'est pas du texte : %v\n", resp.StatusCode)
			continue
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Erreur lors de la lecture de l'URL : %v\n", err)
			continue
		}
		dataStr := string(data)

		p := parser.NewParser(dataStr, urlToCrawl)
		p.Traverse()

		storage, err := db.NewStorage()
		if err != nil {
			continue
		}

		page := db.Page{
			Url:     urlToCrawl,
			Content: p.Content,
			Urls:    p.Url,
		}
		storage.Store(&page)

		for _, url := range p.Url {
			if rb.PathIsAllow(url) && !visited[url] {
				queue = append(queue, url)

			}
		}
	}

}
