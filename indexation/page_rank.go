package indexation

import (
	"log"
	"math"
	"search_egine/db"
)

const (
	dampingFactor = 0.85
	maxIterations = 30
	tolerance     = 0.00001
)

func ComputePageRank() {
	storage, err := db.NewStorage()
	if err != nil {
		return
	}

	pages := db.GetPages(storage, 20000) // Récupère toutes les pages
	if len(pages) == 0 {
		return
	}
	N := float64(len(pages))

	// Initialise les PageRank à 1/N
	pageRanks := make(map[string]float64, len(pages))
	outLinks := make(map[string][]string, len(pages))
	inLinks := make(map[string][]string, len(pages))

	for _, page := range pages {
		pageRanks[page.Url] = 1.0 / N
		for dest := range page.Urls {
			outLinks[page.Url] = append(outLinks[page.Url], dest)
			inLinks[dest] = append(inLinks[dest], page.Url)
		}
	}

	// Boucle d'itération du PageRank
	for i := 0; i < maxIterations; i++ {
		newRanks := make(map[string]float64, len(pages))
		var diff float64

		for _, page := range pages {
			sum := 0.0
			for _, in := range inLinks[page.Url] {
				numOut := float64(len(outLinks[in]))
				if numOut > 0 {
					sum += pageRanks[in] / numOut
				}
			}
			newRank := (1.0-dampingFactor)/N + dampingFactor*sum
			newRanks[page.Url] = newRank
			diff += math.Abs(newRank - pageRanks[page.Url])
		}

		pageRanks = newRanks

		if diff < tolerance {
			break // Convergence atteinte
		}
	}

	// Stockage des résultats via BulkWrite — UN seul appel réseau
	// Avant : N appels UpdateOne séparés (très lent)
	if err := storage.BulkUpdatePageRanks(pageRanks); err != nil {
		log.Printf("Erreur BulkUpdatePageRanks: %v", err)
	}
}
