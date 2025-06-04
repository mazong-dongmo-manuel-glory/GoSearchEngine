package parser

import (
	"golang.org/x/net/html"
	"strings"
)

type Parser struct {
	Content  string     // le contenu brue
	Url      []string   // la liste des urls retourner par le parseur
	RootNode *html.Node // Le noeud racine
	BaseUrl  string
}

// creer un nouveau parser
func NewParser(content string, baseUrl string) *Parser {
	node, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return nil
	}

	return &Parser{
		Content:  "",
		Url:      []string{},
		RootNode: node,
		BaseUrl:  baseUrl,
	}
}

// recuperer toutes les urls
func (p *Parser) Traverse() {
	task := []*html.Node{p.RootNode}
	for len(task) > 0 {
		node := task[len(task)-1]
		task = task[:len(task)-1]

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			// si c'est du texte on l'ajoute au contenu
			if c.Type == html.TextNode && strings.TrimSpace(c.Data) != "" {
				p.Content += strings.TrimSpace(c.Data) + " "
			}

			// si c'est un lien on enregistre tout
			if c.Data == "a" && c.Type == html.ElementNode {
				p.Url = append(p.Url, p.GetAttribute(c, "href"))
			}
			//passer au prochain element
			if c.Type == html.ElementNode {
				task = append(task, c)
			}
		}
	}

}

// Utilitaire pour recuter les attributs
func (p *Parser) GetAttribute(node *html.Node, attrKey string) string {
	attrValue := ""
	if node.Type == html.ElementNode {
		for _, attr := range node.Attr {
			if attr.Key == attrKey {
				attrValue = attr.Val
			}
		}
	}
	return attrValue

}
