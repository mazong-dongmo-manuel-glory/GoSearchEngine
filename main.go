package main

import (
	"fmt"
	"search_egine/crawler"
	"search_egine/indexation"
	"sync"
)

var wg sync.WaitGroup

func main() {
	var sitesFrancophones = []string{
		// Encyclopédies et ressources éducatives
		"https://fr.wikipedia.org",
		"https://fr.wikisource.org",
		"https://www.larousse.fr",
		"https://www.universalis.fr",
		"https://www.cairn.info",
		"https://hal.science",

		// Bibliothèques numériques
		"https://gallica.bnf.fr",
		"https://www.persee.fr",
		"https://www.erudit.org",

		// Sites gouvernementaux et institutionnels
		"https://www.senat.fr",
		"https://www.service-public.fr",
		"https://www.education.gouv.fr",
		"https://www.canada.ca/fr.html",
		"https://www.quebec.ca",

		// Médias avec API ou sitemap
		"https://www.lemonde.fr/sitemap.xml",
		"https://www.lefigaro.fr/sitemap.xml",
		"https://www.liberation.fr/sitemap.xml",
		"https://www.ledevoir.com/sitemap.xml",
		"https://www.lapresse.ca/sitemap.xml",

		// Ressources scientifiques
		"https://www.futura-sciences.com",
		"https://www.pourlascience.fr",
		"https://www.sciencesetavenir.fr",
		"https://www.cirad.fr",

		// Dictionnaires et linguistique
		"https://www.cnrtl.fr",
		"https://dictionnaire.lerobert.com",
		"https://www.dictionnaire-academie.fr",
		"https://www.lexilogos.com",

		// Culture et arts
		"https://www.bnf.fr",
		"https://www.cinematheque.fr",
		"https://www.centrepompidou.fr",
		"https://www.philharmoniedeparis.fr",

		// Ressources francophones canadiennes
		"https://ici.radio-canada.ca",
		"https://www.onf.ca/fr",
		"https://www.banq.qc.ca",

		// Ressources francophones belges et suisses
		"https://www.rtbf.be",
		"https://www.rts.ch",
		"https://www.swissinfo.ch/fre",

		// Archives et patrimoine
		"https://www.archives-nationales.culture.gouv.fr",
		"https://archeologie.culture.fr",
		"https://www.monuments-nationaux.fr",
	}

	queue := crawler.NewQueue()
	queue.AddUrl(sitesFrancophones)
	queue.QueueHandler()
	for i := 0; i < 10; i++ {
		crawler.Wg.Add(1)
		go crawler.CrawlerProcess(i)
	}
	crawler.Wg.Wait()
	fmt.Println("Fin du traitement")
	indexation.ComputePageRank()

}
