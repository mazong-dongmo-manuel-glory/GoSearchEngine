package api

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"search_egine/db"
	"search_egine/indexation"
	"sort"
	"strings"
)

type SearchResult struct {
	PageUrl string  `json:"page_url"`
	Score   float64 `json:"score"`
}

func SearchHandler(c *gin.Context) {
	storage, err := db.NewStorage()
	if err != nil {
		log.Printf("Erreur de base de données : %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Échec de la connexion à la base de données"})
		return
	}
	defer storage.Close()

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Le paramètre de requête 'q' est requis"})
		return
	}

	queryWordsMap := indexation.GetWords(query)
	if len(queryWordsMap) == 0 {
		c.JSON(http.StatusOK, gin.H{"results": []SearchResult{}})
		return
	}

	var queryWords []string
	for word := range queryWordsMap {
		queryWords = append(queryWords, word)
	}

	wordPages, err := storage.GetWordPagesByWords(queryWords, 2000)
	if err != nil {
		log.Printf("Erreur de récupération des WordPages : %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur de récupération des données"})
		return
	}

	pageScores := make(map[string]float64)
	pageBonuses := make(map[string]float64)

	queryLower := strings.ToLower(query)

	for _, wp := range wordPages {
		if _, ok := queryWordsMap[wp.Word]; !ok {
			continue
		}

		pageScores[wp.PageUrl] += wp.Score

		if strings.Contains(strings.ToLower(wp.PageUrl), queryLower) {
			pageBonuses[wp.PageUrl] += 0.5
		}
	}

	var results []SearchResult
	for pageUrl, score := range pageScores {
		score += pageBonuses[pageUrl]
		results = append(results, SearchResult{
			PageUrl: pageUrl,
			Score:   score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	maxResults := 50
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}
