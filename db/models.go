package db

type Page struct {
	Url      string            `bson:"url"`
	Content  string            `bson:"content"`
	Urls     map[string]string `bson:"urls"`
	PageRank float64           `bson:"pagerank"` // ← Ajouté
}

type WordPage struct {
	Word    string  `bson:"word"`
	PageUrl string  `bson:"page_url"`
	TfIdf   float64 `bson:"tfidf"`
	Score   float64 `bson:"score"`
}

func (p *Page) Save() {}
