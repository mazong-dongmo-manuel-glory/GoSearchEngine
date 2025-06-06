package parser

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

type RobotTxt struct {
	DisallowPath map[string]bool
}

func (r *RobotTxt) GetDisallowPath(urlToParse string) {

	r.DisallowPath = make(map[string]bool)
	urlR, err := url.Parse(urlToParse)
	if err != nil {
		return
	}
	robotTxtUrl := urlR.ResolveReference(&url.URL{Path: "/robots.txt"})
	resp, err := http.Get(robotTxtUrl.String())
	if err != nil {
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return
	}
	dataStr := strings.Split(string(data), "\n")
	for _, line := range dataStr {
		splitLine := strings.Split(line, ":")
		if len(splitLine) == 2 {
			if splitLine[0] == "Disallow" {
				r.DisallowPath[strings.TrimSpace(splitLine[1])] = false
			}
		}
	}
}

func (r *RobotTxt) PathIsAllow(urlToParse string) bool {
	for path, _ := range r.DisallowPath {
		if strings.Contains(urlToParse, path) {

			return false
		}
	}
	return true
}
