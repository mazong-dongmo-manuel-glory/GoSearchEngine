package main

import (
	"search_egine/crawler"
)

func main() {

	cr := crawler.Crawler{}
	cr.Crawl("https://ici.radio-canada.ca/")

}
