package db

type Page struct {
	Url     string            `bson:"url"`
	Content string            `bson:"content"`
	Urls    map[string]string `bson:"urls"`
}
type UrlQueuElement struct {
	Url string `bson:"url"`
}

func (p *Page) Save() {}
