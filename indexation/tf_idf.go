package indexation

import (
	"math"
	"search_egine/db"
	"strings"
)

func GetWords(content string) map[string]int {
	words := strings.Split(content, " ")
	var StopWords = map[string]bool{
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

	wordsResult := make(map[string]int)

	for _, word := range words {
		word = strings.TrimSpace(strings.ToLower(word))
		if _, ok := StopWords[word]; ok {
			continue
		}
		if word == "" {
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
	defer storage.Close()

	pages := db.GetPages(storage, 20000)
	if pages == nil || len(pages) == 0 {
		return
	}

	N := float64(len(pages)) // Nombre total de documents

	// 1. Construire DF (Document Frequency)
	df := make(map[string]int)
	pageWords := make([]map[string]int, len(pages)) // pour stocker les mots par page

	for i, page := range pages {
		wordCount := GetWords(page.Content)
		pageWords[i] = wordCount
		seen := make(map[string]bool)

		for word := range wordCount {
			if !seen[word] {
				df[word]++
				seen[word] = true
			}
		}
	}

	// 2. Calculer TF-IDF
	var wordPages []interface{}

	for i, page := range pages {
		wordCount := pageWords[i]

		// Calculer nombre total de mots dans le document
		totalWords := 0
		for _, count := range wordCount {
			totalWords += count
		}

		for word, count := range wordCount {
			tf := float64(count) / float64(totalWords)
			idf := math.Log(N / float64(df[word]))
			tfidf := tf * idf

			wordPages = append(wordPages, db.WordPage{
				Word:    word,
				PageUrl: page.Url,
				TfIdf:   tfidf,
				Score:   tfidf * page.PageRank, // Score initial égal au TF-IDF
			})
		}

		// Insertion par batch si trop gros
		if len(wordPages) >= 1000 {
			storage.StoreMany(wordPages)
			wordPages = []interface{}{}
		}
	}

	// Insertion finale
	if len(wordPages) > 0 {
		storage.StoreMany(wordPages)
	}
}
