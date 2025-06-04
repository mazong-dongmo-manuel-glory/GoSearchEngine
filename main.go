package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

var wg sync.WaitGroup

func main() {
	queue := map[string]bool{}
	fmt.Printf("salut les gens")
	c := colly.NewCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if _, ok := queue[link]; ok {
			delete(queue, link)
			return
		}
		if strings.HasPrefix(link, "http") {
			queue[link] = true
		} else {
			if strings.HasPrefix(link, "/") {
				queue[e.Request.URL.RawQuery+link] = true

			}
		}
		go func() {
			c.Visit(e.Request.URL.RawQuery + link)
			wg.Done()
		}()
		wg.Add(1)
		fmt.Printf("Link found: %s\n", link)
	})
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)

	})

	c.Visit("https://www.google.com")
	wg.Wait()

}
