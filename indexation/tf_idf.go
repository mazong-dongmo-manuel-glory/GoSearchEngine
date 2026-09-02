package indexation

import (
	"math"
	"search_egine/db"
	"strings"
	"unicode"
)

// ── Stop Words — initialisé UNE seule fois au démarrage du programme ──
// Avant : recréé à chaque appel de GetWords() = des millions d'allocations
var stopWords = map[string]bool{
	// Signes de ponctuation
	".":   true,
	",":   true,
	"!":   true,
	"?":   true,
	";":   true,
	":":   true,
	"\"":  true,
	"'":   true,
	"(":   true,
	")":   true,
	"[":   true,
	"]":   true,
	"{":   true,
	"}":   true,
	"-":   true,
	"_":   true,
	"/":   true,
	"\\":  true,
	"|":   true,
	"@":   true,
	"#":   true,
	"$":   true,
	"%":   true,
	"^":   true,
	"&":   true,
	"*":   true,
	"+":   true,
	"=":   true,
	"~":   true,
	"`":   true,
	"<":   true,
	">":   true,
	"«":   true,
	"»":   true,
	"–":   true,
	"—":   true,
	"...": true,

	// Articles français
	"le":    true,
	"la":    true,
	"les":   true,
	"l":     true,
	"un":    true,
	"une":   true,
	"des":   true,
	"du":    true,
	"au":    true,
	"aux":   true,
	"ce":    true,
	"cette": true,
	"ces":   true,
	"mon":   true,
	"ma":    true,
	"mes":   true,
	"ton":   true,
	"ta":    true,
	"tes":   true,
	"son":   true,
	"sa":    true,
	"ses":   true,
	"notre": true,
	"nos":   true,
	"votre": true,
	"vos":   true,
	"leur":  true,
	"leurs": true,

	// Articles anglais
	"the":   true,
	"a":     true,
	"an":    true,
	"some":  true,
	"any":   true,
	"this":  true,
	"that":  true,
	"these": true,
	"those": true,
	"my":    true,
	"your":  true,
	"his":   true,
	"her":   true,
	"its":   true,
	"our":   true,
	"their": true,

	// Prépositions et conjonctions françaises fréquentes
	"de":   true,
	"à":    true,
	"et":   true,
	"en":   true,
	"dans": true,
	"sur":  true,
	"pour": true,
	"avec": true,
	"par":  true,
	"ou":   true,
	"qui":  true,
	"que":  true,
	"dont": true,
	"où":   true,

	// Prépositions et conjonctions anglaises fréquentes
	"of":    true,
	"in":    true,
	"on":    true,
	"at":    true,
	"to":    true,
	"for":   true,
	"with":  true,
	"by":    true,
	"from":  true,
	"and":   true,
	"or":    true,
	"but":   true,
	"which": true,
	"who":   true,
	"what":  true,
	"where": true,
}

// trimPunctuation enlève la ponctuation collée aux mots
// ex: "bonjour," → "bonjour", "(test)" → "test"
func trimPunctuation(word string) string {
	return strings.TrimFunc(word, func(r rune) bool {
		return unicode.IsPunct(r) || unicode.IsSymbol(r)
	})
}

func GetWords(content string) map[string]int {
	words := strings.Fields(content) // Fields est plus rapide que Split(" ") — gère les espaces multiples
	wordsResult := make(map[string]int, len(words)/2)

	for _, word := range words {
		word = strings.ToLower(trimPunctuation(word))
		if word == "" {
			continue
		}
		if stopWords[word] {
			continue
		}
		wordsResult[word]++
	}

	return wordsResult
}

func ProcessTFIDF() {
	storage, err := db.NewStorage()
	if err != nil {
		panic(err)
	}

	pages := db.GetPages(storage, 20000)
	if len(pages) == 0 {
		return
	}

	N := float64(len(pages)) // Nombre total de documents

	// 1. Construire DF (Document Frequency)
	df := make(map[string]int)
	pageWords := make([]map[string]int, len(pages)) // pour stocker les mots par page

	for i, page := range pages {
		wordCount := GetWords(page.Content)
		pageWords[i] = wordCount

		// Chaque mot du document incrémente le DF une seule fois
		for word := range wordCount {
			df[word]++
		}
	}

	// 2. Calculer TF-IDF
	wordPages := make([]interface{}, 0, 1000)

	for i, page := range pages {
		wordCount := pageWords[i]

		// Calculer nombre total de mots dans le document
		totalWords := 0
		for _, count := range wordCount {
			totalWords += count
		}
		if totalWords == 0 {
			continue
		}

		totalWordsF := float64(totalWords)

		for word, count := range wordCount {
			tf := float64(count) / totalWordsF
			idf := math.Log(N / float64(df[word]))
			tfidf := tf * idf

			wordPages = append(wordPages, db.WordPage{
				Word:    word,
				PageUrl: page.Url,
				TfIdf:   tfidf,
				Score:   tfidf * page.PageRank,
			})
		}

		// Insertion par batch si trop gros
		if len(wordPages) >= 1000 {
			storage.StoreMany(wordPages)
			wordPages = wordPages[:0] // Réutilise le slice au lieu d'en allouer un nouveau
		}
	}

	// Insertion finale
	if len(wordPages) > 0 {
		storage.StoreMany(wordPages)
	}
}
