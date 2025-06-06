package parser

import (
	"golang.org/x/net/html"
	"search_egine/utils"
	"strings"
)

type Parser struct {
	Content  string     // le contenu brute
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

// recupere le contenu et les liens
func (p *Parser) Traverse() {
	task := []*html.Node{p.RootNode}
	for len(task) > 0 {
		node := task[len(task)-1]
		task = task[:len(task)-1]

		for c := node.FirstChild; c != nil; c = c.NextSibling {

			if c.Type == html.ElementNode && (c.Data == "script" || c.Data == "style" || c.Data == "iframe" || c.Data == "svg" || c.Data == "img") {
				continue
			}

			// si c'est du texte on l'ajoute au contenu
			if c.Type == html.TextNode && strings.TrimSpace(c.Data) != "" {
				p.Content += strings.TrimSpace(c.Data) + " "
			}

			// si c'est un lien on enregistre tout
			if c.Data == "a" && c.Type == html.ElementNode {
				newUrl := utils.BuildUrl(p.BaseUrl, p.GetAttribute(c, "href"))
				if newUrl != "" {
					p.Url = append(p.Url, newUrl)

				}
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
