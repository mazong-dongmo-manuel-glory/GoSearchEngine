package db

import (
	"net/url"
	"time"
)

type Domains map[string]*Domain

type Domain struct {
	Name        string    `bson:"name"`
	LastCrawlAt time.Time `bson:"last_crawl_at"`
	Urls        []string  `bson:"urls"`
}

type UrlQueuElement struct {
	Url string `bson:"url"`
}
type Page struct {
	Url     string            `bson:"url"`
	Content string            `bson:"content"`
	Urls    map[string]string `bson:"urls"`
}

func NewDomain(urlToParse string) *Domain {
	urlR, err := url.Parse(urlToParse)
	if err != nil {
		return nil
	}
	return &Domain{
		Name:        urlR.Host,
		LastCrawlAt: time.Now(),
		Urls:        []string{urlToParse},
	}
}
func (domains *Domains) AddUrl(urlToParse string) {
	if domains == nil {
		return
	}
	urlR, err := url.Parse(urlToParse)
	if err != nil {
		return
	}
	domainName := urlR.Host
	if _, ok := (*domains)[domainName]; !ok {
		(*domains)[domainName] = NewDomain(urlToParse)
	} else {

		(*domains)[domainName].Urls = append((*domains)[domainName].Urls, urlToParse)
	}

}
func (domains Domains) Size() int {
	return len(domains)
}
func (domains Domains) GetUrlIn() string {
	if domains == nil {
		return ""
	}
	if len(domains) == 0 {
		return ""
	}
	minTime := time.Second * 5

	urlToCrawl := ""
	for _, domain := range domains {
		if domain == nil {
			continue
		}
		domainName := ""
		if time.Now().Sub(domain.LastCrawlAt) > minTime && len(domain.Urls) > 0 {

			domainName = domain.Name
			domain := domains[domainName]
			domain.LastCrawlAt = time.Now()

			urlToCrawl = domain.Urls[0]
			domain.Urls = domain.Urls[1:]
			if len(domain.Urls) == 0 {
				delete(domains, domainName)
			}
			break

		}

	}

	return urlToCrawl

}

func (domain *Domain) Save()   {}
func (domains *Domains) Save() {}
func (p *Page) Save()          {}
