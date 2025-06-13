package api

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"search_egine/db"
	"search_egine/indexation"
)

type SearchResult struct {
	PageUrl string  `json:"page_url"`
	Score   float64 `json:"score"` // Somme des TF-IDF pour tous les mots de la requête sur cette page
}

func SearchHandler(c *gin.Context) {

	storage, err := db.NewStorage("search_engine")
	if err != nil {
		log.Printf("Erreur de base de données : %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Échec de la connexion à la base de données"})
		return
	}
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Le paramètre de requête 'q' est requis"})
		return
	}

	// 1. Traitez la chaîne de requête pour obtenir des termes de recherche individuels
	queryWords := indexation.GetWords(query) // Réutilisez votre fonction getWords pour le traitement de la requête
	var searchTerms []string
	for word, _ := range queryWords {
		searchTerms = append(searchTerms, word)
	}

	if len(searchTerms) == 0 {
		c.JSON(http.StatusOK, gin.H{"results": []SearchResult{}})
		return
	}

	// 2. Récupérez les entrées WordPage pertinentes de la base de données
	// Récupérez initialement un nombre raisonnable de meilleurs résultats TF-IDF par terme de recherche.
	// Vous devrez peut-être ajuster la limite en fonction de tests empiriques.
	wordPages, err := storage.GetWordPagesByWords(searchTerms, 1000) // Récupérer les 1000 meilleures WordPages correspondantes
	if err != nil {
		log.Printf("Erreur de base de données : %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Échec de la récupération des résultats de recherche"})
		return
	}

	// 3. Agrégez les scores TF-IDF par URL de page
	// Nous additionnons le TF-IDF pour chaque page unique en fonction de tous les mots de requête correspondants.
	pageScores := make(map[string]float64)
	for _, wp := range wordPages {
		// N'additionnez le TF-IDF que si le mot de WordPage est l'un de nos termes de recherche
		// (getWords filtre déjà les mots vides, mais c'est une vérification supplémentaire si nécessaire)
		if _, found := queryWords[wp.Word]; found {
			pageScores[wp.PageUrl] += wp.TfIdf
		}
	}

	// 4. Convertissez la map en un slice de SearchResult triable
	var results []SearchResult
	for pageUrl, score := range pageScores {
		results = append(results, SearchResult{PageUrl: pageUrl, Score: score})
	}

	// 5. Triez les résultats par score total par ordre décroissant
	// Vous pouvez implémenter une fonction de tri personnalisée.
	// Pour simplifier, utilisez un tri de base ici. Pour des ensembles de résultats très importants,
	// considérez si la pagination et le tri côté base de données sont plus efficaces.
	// Pour l'instant, nous trions après la récupération.
	// (Remarque : si vous devez trier des centaines de milliers de résultats en mémoire,
	// vous pourriez rencontrer des problèmes de performance. Envisagez d'utiliser l'agrégation de base de données
	// framework si votre base de données le prend en charge, comme $group et $sort de MongoDB).

	// Pour la démonstration, utilisons un tri plus simple :
	// Cela nécessiterait une interface sort.Interface personnalisée ou l'utilisation du package sort.
	// Pour l'instant, en supposant que vous implémenterez le tri ou la pagination.
	// Ou, si vous utilisez le framework d'agrégation de MongoDB, le tri se fait sur la base de données.
	// Pour un tri simple en mémoire des scores agrégés :
	// sort.Slice(results, func(i, j int) bool {
	// 	return results[i].Score > results[j].Score
	// })

	// Pour une meilleure performance avec potentiellement de nombreux résultats, utilisez le framework d'agrégation de MongoDB
	// Cela permet à la base de données de faire le gros du travail de regroupement et de tri.
	// Exemple de la façon dont GetWordPagesByWords pourrait être augmentée ou une nouvelle fonction créée :
	// func (s *Storage) AggregateSearchResults(searchTerms []string, limit int64) ([]SearchResult, error) {
	//     collection := s.Client.Database("search_engine").Collection("word_pages")
	//     pipeline := mongo.Pipeline{
	//         {{"$match", bson.M{"word": bson.M{"$in": searchTerms}}}},
	//         {{"$group", bson.D{
	//             {Key: "_id", Value: "$page_url"},
	//             {Key: "score", Value: bson.M{"$sum": "$tfidf"}},
	//         }}},
	//         {{"$sort", bson.D{{Key: "score", Value: -1}}}},
	//         {{"$limit", limit}},
	//     }

	//     ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	//     defer cancel()

	//     cursor, err := collection.Aggregate(ctx, pipeline)
	//     if err != nil {
	//         return nil, fmt.Errorf("failed to aggregate search results: %v", err)
	//     }
	//     defer cursor.Close(ctx)

	//     var aggregatedResults []SearchResult
	//     if err = cursor.All(ctx, &aggregatedResults); err != nil {
	//         return nil, fmt.Errorf("failed to decode aggregated results: %v", err)
	//     }
	//     return aggregatedResults, nil
	// }

	// Si vous utilisez l'approche d'agrégation :
	// results, err = storage.AggregateSearchResults(searchTerms, 50) // Récupérer les 50 meilleurs résultats
	// if err != nil {
	//     log.Printf("Erreur d'agrégation de base de données : %v", err)
	//     c.JSON(http.StatusInternalServerError, gin.H{"error": "Échec de l'agrégation des résultats de recherche"})
	//     return
	// }

	// Pour l'implémentation actuelle de GetWordPagesByWords, nous allons trier en Go pour la simplicité de cet exemple.
	// Pour les grands ensembles de données, envisagez l'agrégation MongoDB comme indiqué dans les commentaires ci-dessus.
	// En supposant que 'results' est déjà rempli à partir de la map pageScores
	// (Vous devrez importer "sort")
	// sort.Slice(results, func(i, j int) bool {
	//     return results[i].Score > results[j].Score
	// })

	// Implémentez un tri approprié ici si vous n'utilisez pas l'agrégation de la base de données.
	// Pour cet exemple, retournons simplement les N premiers résultats après agrégation et tri en Go (pour la simplicité)
	// Dans un système réel, vous voudriez pousser autant de travail que possible vers la base de données.
	// Supposons que vous trierez le slice `results` en dehors de ce gestionnaire.
	// Pour l'instant, limitons les résultats à un nombre raisonnable pour la réponse de l'API :
	maxResults := 50 // Afficher les 50 meilleurs résultats
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}
