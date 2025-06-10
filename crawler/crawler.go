package crawler

import (
	"fmt"
	"io"
	"net/http"
	url2 "net/url"
	"search_egine/db"
	"search_egine/parser"
	"strings"
	"sync"
	"time"
)

const MinTimeBetweenRequest = 10 * time.Second
const MaxIterationForGetUrl = 1000

var urlChanSender = make(chan string, 100)
var urlChanReceiver = make(chan []string, 100)
var Wg = &sync.WaitGroup{}

type Queue struct {
	Urls    []string
	Visited map[string]interface{}
	Domains map[string]*Domain
	mu      sync.Mutex
}
type Domain struct {
	RobotTxt        *parser.RobotTxt
	LastVisitedTime time.Time
}

func NewQueue() *Queue {
	return &Queue{
		Urls:    []string{},
		Visited: make(map[string]interface{}),
		Domains: make(map[string]*Domain),
		mu:      sync.Mutex{},
	}
}

func (q *Queue) AddUrl(urls []string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, url := range urls {
		if _, ok := q.Visited[url]; ok {
			continue
		}
		urlParsed, err := url2.Parse(url)
		if err != nil {
			continue
		}
		if _, ok := q.Domains[urlParsed.Host]; !ok {
			q.Domains[urlParsed.Host] = &Domain{
				RobotTxt:        &parser.RobotTxt{},
				LastVisitedTime: time.Now(),
			}
		}

		q.Urls = append(q.Urls, url)
	}

}

// GetUrl corrigée pour récupérer une URL unique et prête à être crawlée
func (q *Queue) GetUrl() string {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Itérer un nombre limité de fois pour trouver une URL appropriée
	// Ceci évite de boucler indéfiniment si toutes les URLs sont temporairement non-crawlables
	for i := 0; i < MaxIterationForGetUrl; i++ {
		// Si la file est vide, rien à récupérer pour le moment
		if len(q.Urls) == 0 {
			// Pas de temps de sommeil ici, le code appelant (ex: QueueHandler) gérera l'attente
			return ""
		}

		// Récupérer l'URL candidate du début de la file
		currentURL := q.Urls[0]

		// 1. Vérifier si l'URL a déjà été visitée (ou est en cours de traitement ailleurs)
		// (Idéalement, AddUrl devrait déjà empêcher les duplicatas, mais cette re-vérification est une sécurité)
		if _, ok := q.Visited[currentURL]; ok {
			// Si déjà visitée, la déplacer à la fin de la file et passer à l'URL suivante
			q.Urls = q.Urls[1:]
			q.Urls = append(q.Urls, currentURL)
			continue // Passer à l'itération suivante de la boucle
		}

		// 2. Analyser l'URL pour extraire le domaine (nécessaire pour la limitation de débit)
		urlParsed, err := url2.Parse(currentURL)
		if err != nil {
			fmt.Printf("GetUrl: Erreur d'analyse d'URL '%s': %v\n", currentURL, err)
			// Si l'URL est malformée, la déplacer à la fin de la file et passer à l'URL suivante
			q.Urls = q.Urls[1:]
			q.Urls = append(q.Urls, currentURL)
			continue // Passer à l'itération suivante de la boucle
		}

		host := urlParsed.Host
		// Initialiser le domaine si c'est la première fois que nous le rencontrons
		if q.Domains[host] == nil {
			q.Domains[host] = &Domain{
				LastVisitedTime: time.Time{}, // Temps zéro, indique qu'il n'a jamais été visité
			}
		}

		// 3. Vérifier la limitation de débit pour ce domaine
		// On ne vérifie que si le domaine a déjà été visité (LastVisitedTime n'est pas le temps zéro)
		if !q.Domains[host].LastVisitedTime.IsZero() && time.Now().Sub(q.Domains[host].LastVisitedTime) < MinTimeBetweenRequest {
			// Si le délai minimum n'est pas écoulé, déplacer l'URL à la fin de la file et passer à l'URL suivante

			q.Urls = append(q.Urls, currentURL)
			continue // Passer à l'itération suivante de la boucle
		}
		if q.Domains[host].RobotTxt.CheckIfIsDisAllowPath(currentURL) {
			q.Urls = q.Urls[1:]
			continue
		}
		// Si toutes les vérifications passent, cette URL est prête à être retournée
		// Mettre à jour le temps de la dernière visite pour ce domaine
		q.Domains[host].LastVisitedTime = time.Now()
		// Retirer l'URL de la file d'attente (car elle est maintenant "prise")
		q.Urls = q.Urls[1:]
		q.Visited[currentURL] = nil // Marquer l'URL comme visitée'

		return currentURL // Retourner l'URL valide
	}

	// Si la boucle se termine sans trouver d'URL appropriée (ex: toutes sont limitées ou invalides)
	// Ne pas bloquer ici, laisser l'appelant décider quoi faire.
	return "" // Aucune URL appropriée trouvée pour le moment
}
func (q *Queue) QueueHandler() {
	Wg.Add(2)
	go func() {
		defer Wg.Done()
		for {
			url := q.GetUrl()
			urlChanSender <- url
		}
	}()
	go func() {
		defer Wg.Done()
		for {
			url := <-urlChanReceiver
			q.AddUrl(url)
		}

	}()

}

func CrawlerProcess(id int) {
	Wg.Add(1)
	defer Wg.Done()
	storage, err := db.NewStorage("search_engine")
	if err != nil {
		fmt.Printf("Error %v\n", err)
		return
	}
	defer storage.Close()

	for {
		select {
		case url := <-urlChanSender:
			if url == "" {
				continue
			}
			fmt.Printf("Crawler %d : %s\n", id, url)
			resp, err := http.Get(url)
			if err != nil {

				continue
			}
			if resp.StatusCode != 200 || !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
				resp.Body.Close()
				continue
			}
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				resp.Body.Close()
				continue
			}
			dataStr := string(data)
			p := parser.NewParser(dataStr, url)
			p.Traverse()
			storage.Store(p)
			urls := make([]string, 0)
			for url, _ := range p.Url {
				urls = append(urls, url)
			}
			urlChanReceiver <- urls
		}
	}

}
