package parser

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

type RobotTxt struct {
	DisAllowPath map[string]string
}

func (rb *RobotTxt) GetDisallowPath(urlToFind string) {
	urlToFindParsed, err := url.Parse(urlToFind)
	rb.DisAllowPath = make(map[string]string)
	if err != nil {
		return
	}

	robotUrl := urlToFindParsed.ResolveReference(&url.URL{Path: "/robots.txt"})

	resp, err := http.Get(robotUrl.String())
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 || !strings.Contains(resp.Header.Get("Content-Type"), "text/plain") {
		return
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	dataStr := string(data)
	prevUserAgent := make(map[string]string)
	endOfNewUserAgent := false
	for _, line := range strings.Split(dataStr, "\n") {
		if strings.Contains(line, "User-agent") && !endOfNewUserAgent {
			prevUserAgent[line] = ""
		} else if strings.Contains(line, "User-agent") && endOfNewUserAgent {
			endOfNewUserAgent = false
			prevUserAgent = make(map[string]string)
			prevUserAgent[line] = ""

		} else {
			endOfNewUserAgent = true
			if strings.Contains(line, "Disallow") {
				lines := strings.Split(line, ":")
				_, ok := prevUserAgent["User-agent: *"]
				if len(lines) > 1 && ok {
					rb.DisAllowPath[strings.TrimSpace(lines[1])] = ""
				}
			}
		}

	}

}

// retourne true si l'url est dans la liste des disallow path'
func (rb *RobotTxt) CheckIfIsDisAllowPath(urlToFind string) bool {
	for urlR, _ := range rb.DisAllowPath {
		if strings.Contains(urlToFind, urlR) {
			return true
		}
	}
	return false
}
