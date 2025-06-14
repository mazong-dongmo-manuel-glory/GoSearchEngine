package indexation

import (
	"strings"
)

const (
	BATCH_SIZE       = 1000  // Traiter 1000 pages à la fois
	STORE_BATCH_SIZE = 10000 // Stocker 10000 WordPages à la fois
)

// Structure pour stocker les statistiques globales
type GlobalStats struct {
	TotalPages     int
	WordDocCount   map[string]int // Nombre de documents contenant chaque mot
	WordTotalCount map[string]int // Nombre total d'occurrences de chaque mot
}

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
