package utils

import (
	"net/url"
	"path"
	"strings"
)

func BuildUrl(baseUrl string, newUrl string) string {
	// 1. Vérifications preliminarily
	if newUrl == "" {
		baseUrlObj, err := url.Parse(baseUrl)
		if err != nil {
			return ""
		}
		baseUrlObj.Fragment = ""
		return baseUrlObj.String()
	}

	// 2. Treatment des URLs non-http
	if strings.HasPrefix(newUrl, "mailto:") ||
		strings.HasPrefix(newUrl, "ftp:") ||
		strings.HasPrefix(newUrl, "tel:") {
		return ""
	}

	// 3. Traitement des fragments seuls
	if strings.HasPrefix(newUrl, "#") {
		baseUrlObj, err := url.Parse(baseUrl)
		if err != nil {
			return ""
		}
		baseUrlObj.Fragment = ""
		return baseUrlObj.String()
	}

	// 4. Parse des URLs
	baseUrlObj, err := url.Parse(baseUrl)
	if err != nil {
		return ""
	}

	// 5. Déterminer le répertoire de base
	baseDir := path.Dir(baseUrlObj.Path)
	if !strings.HasSuffix(baseDir, "/") {
		baseDir += "/"
	}

	// 6. Traiter la nouvelle URL
	newUrlObj, err := url.Parse(newUrl)
	if err != nil {
		return ""
	}

	// 7. Nettoyage des fragments
	newUrlObj.Fragment = ""

	// 8. Traitement des URLs relatives au protocole (//example.com)
	if strings.HasPrefix(newUrl, "//") {
		return "https:" + newUrl
	}

	// 9. Traitement des URLs absolues
	if newUrlObj.IsAbs() {
		if newUrlObj.Scheme != "http" && newUrlObj.Scheme != "https" {
			return ""
		}
		return newUrlObj.String()
	}

	// 10. Traitement des domaines nus (example.com)
	if !strings.HasPrefix(newUrl, "/") &&
		!strings.HasPrefix(newUrl, ".") &&
		strings.Contains(newUrl, ".") &&
		!strings.Contains(newUrl, "/") &&
		!strings.HasSuffix(newUrl, ".html") &&
		!strings.HasSuffix(newUrl, ".htm") &&
		!strings.HasSuffix(newUrl, ".php") &&
		!strings.HasSuffix(newUrl, ".png") &&
		!strings.HasSuffix(newUrl, ".jpg") &&
		!strings.HasSuffix(newUrl, ".gif") {
		return "https://" + newUrl
	}

	// 11. Construction de l'URL finale pour les chemins relatifs
	var finalUrl *url.URL
	if strings.HasPrefix(newUrl, "/") {
		// Chemin absolu
		finalUrl = &url.URL{
			Scheme: baseUrlObj.Scheme,
			Host:   baseUrlObj.Host,
			Path:   newUrl,
		}
	} else if strings.HasPrefix(newUrl, "../") {
		// Remonter d'un niveau
		parentDir := path.Dir(path.Dir(baseUrlObj.Path))
		if !strings.HasSuffix(parentDir, "/") {
			parentDir += "/"
		}
		finalUrl = &url.URL{
			Scheme: baseUrlObj.Scheme,
			Host:   baseUrlObj.Host,
			Path:   path.Join(parentDir, strings.TrimPrefix(newUrl, "../")),
		}
	} else {
		// Chemin relatif dans le même dossier
		finalUrl = &url.URL{
			Scheme: baseUrlObj.Scheme,
			Host:   baseUrlObj.Host,
			Path:   path.Join(baseDir, newUrl),
		}
	}

	// 12. Gestion du protocole
	if strings.Contains(finalUrl.Host, "localhost") {
		finalUrl.Scheme = baseUrlObj.Scheme
	} else if finalUrl.Host != baseUrlObj.Host {
		finalUrl.Scheme = "https"
	}

	return finalUrl.String()
}
