package indexation

import (
	"fmt"
	"math"
	"search_egine/db"
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

// ÉTAPE 1: Calculer les statistiques globales sur toute la collection
func calculateGlobalStats(storage *db.Storage, totalPages int) (*GlobalStats, error) {
	fmt.Println("=== ÉTAPE 1: Calcul des statistiques globales ===")

	stats := &GlobalStats{
		TotalPages:     totalPages,
		WordDocCount:   make(map[string]int),
		WordTotalCount: make(map[string]int),
	}

	processed := 0

	for offset := 0; offset < totalPages; offset += BATCH_SIZE {
		limit := BATCH_SIZE
		if offset+BATCH_SIZE > totalPages {
			limit = totalPages - offset
		}

		pages := db.GetPagesWithOffset(storage, limit, offset)
		if len(pages) == 0 {
			break
		}

		for _, page := range pages {
			words := GetWords(page.Content)
			uniqueWordsInPage := make(map[string]bool)

			// Compter les occurrences totales et marquer les mots uniques par page
			for word, count := range words {
				stats.WordTotalCount[word] += count
				uniqueWordsInPage[word] = true
			}

			// Compter le nombre de documents contenant chaque mot
			for word := range uniqueWordsInPage {
				stats.WordDocCount[word]++
			}
		}

		processed += len(pages)
		if processed%5000 == 0 {
			fmt.Printf("Pages analysées : %d/%d (%.1f%%)\n",
				processed, totalPages, float64(processed)/float64(totalPages)*100)
		}
	}

	fmt.Printf("Statistiques globales calculées :\n")
	fmt.Printf("- Total pages : %d\n", stats.TotalPages)
	fmt.Printf("- Mots uniques : %d\n", len(stats.WordDocCount))
	fmt.Printf("- Occurrences totales : %d\n", func() int {
		total := 0
		for _, count := range stats.WordTotalCount {
			total += count
		}
		return total
	}())

	return stats, nil
}

// ÉTAPE 2: Calculer TF-IDF avec les statistiques globales
func calculateTfIdfWithGlobalStats(storage *db.Storage, totalPages int, stats *GlobalStats) error {
	fmt.Println("=== ÉTAPE 2: Calcul TF-IDF avec statistiques globales ===")

	var batch []interface{}
	processed := 0
	totalWordPages := 0

	for offset := 0; offset < totalPages; offset += BATCH_SIZE {
		limit := BATCH_SIZE
		if offset+BATCH_SIZE > totalPages {
			limit = totalPages - offset
		}

		pages := db.GetPagesWithOffset(storage, limit, offset)
		if len(pages) == 0 {
			break
		}

		for _, page := range pages {
			words := GetWords(page.Content)
			totalWordsInPage := len(words)

			if totalWordsInPage == 0 {
				continue
			}

			for word, count := range words {
				// Vérifier que le mot existe dans les statistiques globales
				docFreq, exists := stats.WordDocCount[word]
				if !exists || docFreq == 0 {
					continue // Ignorer les mots non trouvés (ne devrait pas arriver)
				}

				wordPage := &db.WordPage{
					Word:    word,
					PageUrl: page.Url,
					Score:   page.PageRank,
				}

				// Calcul TF (Term Frequency) - fréquence relative dans le document
				tf := float64(count) / float64(totalWordsInPage)

				// Calcul IDF (Inverse Document Frequency) - basé sur toute la collection
				idf := math.Log(float64(stats.TotalPages) / float64(docFreq))

				// Calcul TF-IDF
				tfIdf := tf * idf

				// Score final : combinaison TF-IDF et PageRank
				// Le score reflète maintenant correctement la rareté du mot dans toute la collection
				wordPage.TfIdf = 0.7*math.Abs(tfIdf) + 0.3*wordPage.Score

				batch = append(batch, wordPage)
				totalWordPages++

				// Stockage par lots
				if len(batch) >= STORE_BATCH_SIZE {
					storage.StoreMany(batch)

					batch = batch[:0]
					fmt.Printf("Stocké %d WordPages (total: %d)\n", STORE_BATCH_SIZE, totalWordPages)
				}
			}
		}

		processed += len(pages)
		if processed%1000 == 0 {
			fmt.Printf("Pages indexées : %d/%d (%.1f%%)\n",
				processed, totalPages, float64(processed)/float64(totalPages)*100)
		}
	}

	// Stocker le dernier lot
	if len(batch) > 0 {
		storage.StoreMany(batch)

		fmt.Printf("Stocké les %d dernières WordPages\n", len(batch))
	}

	fmt.Printf("=== INDEXATION TERMINÉE ===\n")
	fmt.Printf("Total WordPages créées : %d\n", totalWordPages)
	return nil
}

// Fonction principale d'indexation optimisée
func Indexation() {
	fmt.Println("=== DÉBUT INDEXATION OPTIMISÉE ===")

	storage, err := db.NewStorage("search_engine")
	if err != nil {
		fmt.Printf("Erreur lors de l'ouverture du storage : %v\n", err)
		return
	}
	defer storage.Close()

	// Compter le nombre total de pages
	totalPages := db.CountPages(storage)
	if totalPages == 0 {
		fmt.Println("Aucune page à indexer")
		return
	}

	fmt.Printf("Indexation de %d pages au total\n", totalPages)

	// ÉTAPE 1: Calculer les statistiques globales
	stats, err := calculateGlobalStats(storage, totalPages)
	if err != nil {
		fmt.Printf("Erreur lors du calcul des statistiques globales : %v\n", err)
		return
	}

	// ÉTAPE 2: Calculer TF-IDF avec les vraies statistiques globales
	err = calculateTfIdfWithGlobalStats(storage, totalPages, stats)
	if err != nil {
		fmt.Printf("Erreur lors du calcul TF-IDF : %v\n", err)
		return
	}

	// Afficher quelques statistiques finales
	displayFinalStats(stats)

	fmt.Println("=== INDEXATION RÉUSSIE ===")
}

// Afficher des statistiques utiles pour validation
func displayFinalStats(stats *GlobalStats) {
	fmt.Println("\n=== STATISTIQUES FINALES ===")

	// Trouver les mots les plus fréquents
	type WordStat struct {
		Word       string
		DocCount   int
		TotalCount int
	}

	var topWords []WordStat
	for word, docCount := range stats.WordDocCount {
		if len(topWords) < 10 {
			topWords = append(topWords, WordStat{word, docCount, stats.WordTotalCount[word]})
		} else {
			// Remplacer le moins fréquent si nécessaire
			minIdx := 0
			for i, w := range topWords {
				if w.DocCount < topWords[minIdx].DocCount {
					minIdx = i
				}
			}
			if docCount > topWords[minIdx].DocCount {
				topWords[minIdx] = WordStat{word, docCount, stats.WordTotalCount[word]}
			}
		}
	}

	fmt.Println("Top 10 mots les plus fréquents :")
	for i, w := range topWords {
		if i < len(topWords) {
			idf := math.Log(float64(stats.TotalPages) / float64(w.DocCount))
			fmt.Printf("%d. '%s' - Documents: %d (%.1f%%) - Total: %d - IDF: %.3f\n",
				i+1, w.Word, w.DocCount,
				float64(w.DocCount)/float64(stats.TotalPages)*100,
				w.TotalCount, idf)
		}
	}

	// Statistiques sur la distribution
	fmt.Printf("\nDistribution des mots :\n")
	rare := 0     // mots dans < 1% des documents
	common := 0   // mots dans 1-10% des documents
	frequent := 0 // mots dans > 10% des documents

	for _, docCount := range stats.WordDocCount {
		percentage := float64(docCount) / float64(stats.TotalPages) * 100
		if percentage < 1 {
			rare++
		} else if percentage < 10 {
			common++
		} else {
			frequent++
		}
	}

	fmt.Printf("- Mots rares (< 1%% docs) : %d\n", rare)
	fmt.Printf("- Mots communs (1-10%% docs) : %d\n", common)
	fmt.Printf("- Mots fréquents (> 10%% docs) : %d\n", frequent)
}

// Version alternative pour validation - échantillonnage
func ValidateIndexation(sampleSize int) {
	fmt.Printf("=== VALIDATION SUR %d PAGES ===\n", sampleSize)

	storage, err := db.NewStorage("search_engine")
	if err != nil {
		fmt.Printf("Erreur : %v\n", err)
		return
	}
	defer storage.Close()

	pages := db.GetPagesWithOffset(storage, sampleSize, 0)

	// Calculer sur l'échantillon
	stats, _ := calculateGlobalStats(storage, len(pages))

	fmt.Println("Validation terminée - vérifiez la cohérence des scores")
	displayFinalStats(stats)
}

// Ajouter ces fonctions à votre package db
