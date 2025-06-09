package crawler

import (
	url2 "net/url"
	"time"
)

var queue = Queue{}

const MinTimeBetweenCrawl = 1 * time.Second

type Domains map[string]time.Time
type Queue []string

func (d Domains) Size() int {
	return len(d)
}
func (q *Queue) Get() string {
	if len(*q) == 0 {
		return ""
	}
	url := (*q)[0]
	*q = (*q)[1:]
	return url

}
func getHostname(url string) string {
	urlParsed, err := url2.Parse(url)
	if err != nil {
		return ""
	}
	return urlParsed.Hostname()
}
func AddNewUrlInQueue(url string) {

	domainOfUrlParsed := getHostname(url)
	domains[domainOfUrlParsed] = time.Now()
	queue = append(queue, url)

}
func GetUrlInQueue() string {
	initialQueueLength := len(queue)

	for i := 0; i < initialQueueLength; i++ {
		url := queue.Get()
		if url == "" {
			return ""
		}

		domainOfUrl := getHostname(url)
		if domainOfUrl == "" {
			continue
		}

		timeOfLastGet, ok := domains[domainOfUrl]
		if ok && time.Now().Sub(timeOfLastGet) < MinTimeBetweenCrawl {
			// Trop récent, on le remet à la fin
			queue = append(queue, url)
			continue
		} else {
			domains[domainOfUrl] = time.Now()
			return url
		}
		// Ok pour traitement
	}

	// Aucun URL n'est disponible
	return ""
}
