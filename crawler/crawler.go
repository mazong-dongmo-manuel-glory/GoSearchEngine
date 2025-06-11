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

// Délai minimum entre deux requêtes vers un même domaine
const MinTimeBetweenRequest = 40 * time.Second

// Nombre maximal d'itérations dans la boucle de GetUrl
const MaxIterationForGetUrl = 100000

// Taille maximale de la file d'attente (URLs)
const MaxSizeQueue = 1000000

// Canal pour envoyer une URL au processus de crawl
var urlChanSender = make(chan string, 1)

// Canal pour recevoir une liste d'URLs à ajouter à la queue
var urlChanReceiver = make(chan []string, 1)

// WaitGroup global pour la synchronisation des goroutines
var Wg = &sync.WaitGroup{}

// Structure principale représentant la file d'attente d'URLs à crawler
type Queue struct {
	Urls    []string               // Liste des URLs à traiter
	Visited map[string]interface{} // URLs déjà visitées
	Domains map[string]*Domain     // Infos par domaine (robots.txt, derniers accès)
	mu      sync.Mutex             // Mutex pour protéger l'accès concurrent à la file
	start   time.Time              // Temps de démarrage (pour gérer certains délais)
}

// Infos sur un domaine (robots.txt et dernier accès)
type Domain struct {
	RobotTxt        *parser.RobotTxt
	LastVisitedTime time.Time
}

// Initialise une nouvelle instance de la queue
func NewQueue() *Queue {
	return &Queue{
		Urls:    []string{},
		Visited: make(map[string]interface{}),
		Domains: make(map[string]*Domain),
		mu:      sync.Mutex{},
		start:   time.Now(),
	}
}

// Ajoute des URLs à la file d'attente
func (q *Queue) AddUrl(urls []string) {
	q.mu.Lock()
	defer func() {
		// Nettoyage de la queue si elle dépasse la taille maximale
		if len(q.Urls) > MaxSizeQueue {
			q.Urls = q.Urls[MaxSizeQueue-10000:]
			q.Visited = make(map[string]interface{}) // Réinitialise les URLs visitées
			q.Domains = make(map[string]*Domain)     // Réinitialise les domaines
		}
		q.mu.Unlock()
	}()

	for _, url := range urls {
		if _, ok := q.Visited[url]; ok {
			continue // Ignore les URLs déjà visitées
		}
		urlParsed, err := url2.Parse(url)
		if err != nil {
			continue // Ignore les URLs invalides
		}
		// Initialise le domaine si non présent
		if _, ok := q.Domains[urlParsed.Host]; !ok {
			q.Domains[urlParsed.Host] = &Domain{
				RobotTxt:        &parser.RobotTxt{},
				LastVisitedTime: time.Now(),
			}
		}
		// Ajoute l'URL à la queue
		q.Urls = append(q.Urls, url)
	}
}

// Récupère une URL à crawler, en respectant robots.txt et les délais
func (q *Queue) GetUrl() string {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.Urls) == 0 {
		return ""
	}

	url := ""
	for i := 0; i < MaxIterationForGetUrl && i < len(q.Urls); i++ {
		url = q.Urls[0]

		// Ignore si déjà visitée
		if _, ok := q.Visited[url]; ok {
			q.Urls = q.Urls[1:]
			continue
		}

		urlParsed, err := url2.Parse(q.Urls[0])
		if err != nil {
			q.Urls = q.Urls[1:]
			continue
		}

		// Si le domaine existe, vérifie les délais et les règles robots.txt
		if d, ok := q.Domains[urlParsed.Host]; ok {
			// Si le dernier accès est trop récent et qu'on a attendu un peu depuis le début
			if time.Now().Sub(d.LastVisitedTime) < MinTimeBetweenRequest && time.Now().Sub(q.start) > time.Second*10 {
				// On repousse l'URL à la fin de la queue
				q.Urls = q.Urls[1:]
				q.Urls = append(q.Urls, url)
				continue
			}
			// Si robots.txt bloque ce chemin
			if d.RobotTxt.CheckIfIsDisAllowPath(url) {
				q.Urls = q.Urls[1:]
				continue
			}
		}

		// Si domaine inconnu (ex: nettoyé précédemment), on l'initialise
		if _, ok := q.Domains[urlParsed.Host]; !ok {
			q.Domains[urlParsed.Host] = &Domain{
				RobotTxt:        &parser.RobotTxt{},
				LastVisitedTime: time.Now(),
			}
			// Récupère les règles de robots.txt
			q.Domains[urlParsed.Host].RobotTxt.GetDisallowPath(url)
		}

		// Supprime l'URL de la queue (prête à être traitée)
		q.Urls = q.Urls[1:]

		// Met à jour le temps de dernière visite pour le domaine
		q.Domains[urlParsed.Host].LastVisitedTime = time.Now()
		// Marque l'URL comme visitée
		q.Visited[url] = nil
		return url
	}
	// Retourne la dernière URL considérée même si elle est invalide (à corriger éventuellement)
	return url
}

// Lance deux goroutines : l'une pour envoyer des URLs à traiter, l'autre pour en recevoir et les ajouter
func (q *Queue) QueueHandler() {
	Wg.Add(2)

	// Goroutine pour envoyer une URL à traiter
	go func() {
		defer Wg.Done()
		for {
			url := q.GetUrl()
			urlChanSender <- url
		}
	}()

	// Goroutine pour ajouter de nouvelles URLs reçues
	go func() {
		defer Wg.Done()
		for {
			url := <-urlChanReceiver
			q.AddUrl(url)
		}
	}()
}

// Fonction exécutée par un worker crawler (identifié par un ID)
func CrawlerProcess(id int) {
	storage, err := db.NewStorage("search_engine")
	if err != nil {
		return
	}

	Wg.Add(1)
	defer Wg.Done()

	for {
		// Reçoit une URL à crawler
		url := <-urlChanSender
		fmt.Printf("Crawler %d : %s\n", id, url)

		// Envoie une requête HTTP GET
		resp, err := http.Get(url)
		if err != nil {
			continue // Ignore en cas d’erreur réseau
		}

		// Ignore si le contenu n'est pas HTML
		if resp.StatusCode != 200 || !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
			resp.Body.Close()
			continue
		}

		// Lit le contenu de la réponse
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			continue
		}

		// Ferme le corps de la réponse HTTP
		dataStr := string(data)
		p := parser.NewParser(dataStr, url)

		// Analyse le contenu HTML
		p.Traverse()

		// Stocke la page dans la base de données
		page := db.Page{
			Url:     url,
			Content: p.Content,
			Urls:    p.Url,
		}
		storage.Store(&page)

		// Extrait les URLs trouvées sur la page
		urls := make([]string, 0)
		for url := range p.Url {
			urls = append(urls, url)
		}

		// Envoie les nouvelles URLs au gestionnaire de queue
		urlChanReceiver <- urls
	}
}
