package db

type Page struct {
	Url     string   `bson:"url"`
	Content string   `bson:"content"`
	Urls    []string `bson:"urls"`
}

func (p *Page) Save() {

}
