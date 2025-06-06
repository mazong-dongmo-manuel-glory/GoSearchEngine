package url_test

import (
	"search_egine/utils"
	"testing"
)

func TestBuildUrl(t *testing.T) {
	baseUrl := "https://example.com/current/page.html"
	tests := []struct {
		name    string
		baseUrl string
		newUrl  string
		want    string
	}{
		// 1. URLs absolues
		{
			name:    "URL HTTPS absolue",
			baseUrl: baseUrl,
			newUrl:  "https://anothersite.com/page.html",
			want:    "https://anothersite.com/page.html",
		},
		{
			name:    "URL HTTP absolue",
			baseUrl: baseUrl,
			newUrl:  "http://anothersite.com/page.html",
			want:    "http://anothersite.com/page.html",
		},

		// 2. URLs relatives au root
		{
			name:    "Chemin absolu depuis la racine",
			baseUrl: baseUrl,
			newUrl:  "/path/to/resource",
			want:    "https://example.com/path/to/resource",
		},
		{
			name:    "Chemin court depuis la racine",
			baseUrl: baseUrl,
			newUrl:  "/a",
			want:    "https://example.com/a",
		},

		// 3. URLs relatives
		{
			name:    "Chemin relatif simple",
			baseUrl: "https://example.com/base/",
			newUrl:  "another/resource",
			want:    "https://example.com/base/another/resource",
		},
		{
			name:    "Chemin relatif avec parent directory",
			baseUrl: "https://example.com/base/current/",
			newUrl:  "../sibling/resource.html",
			want:    "https://example.com/base/sibling/resource.html",
		},
		{
			name:    "Fichier dans le même dossier",
			baseUrl: "https://example.com/folder/page.html",
			newUrl:  "image.png",
			want:    "https://example.com/folder/image.png",
		},

		// 4. Cas spéciaux
		{
			name:    "URL protocol-relative",
			baseUrl: baseUrl,
			newUrl:  "//example.org/path",
			want:    "https://example.org/path",
		},
		{
			name:    "Domaine nu",
			baseUrl: baseUrl,
			newUrl:  "google.com",
			want:    "https://google.com",
		},
		{
			name:    "URL vide",
			baseUrl: baseUrl,
			newUrl:  "",
			want:    "https://example.com/current/page.html",
		},
		{
			name:    "Fragment seul",
			baseUrl: baseUrl,
			newUrl:  "#top",
			want:    "https://example.com/current/page.html",
		},

		// 5. URLs non supportées
		{
			name:    "URL FTP",
			baseUrl: baseUrl,
			newUrl:  "ftp://ftp.example.com/file.zip",
			want:    "",
		},
		{
			name:    "URL mailto",
			baseUrl: baseUrl,
			newUrl:  "mailto:test@example.com",
			want:    "",
		},
		{
			name:    "URL invalide",
			baseUrl: baseUrl,
			newUrl:  "::::invalid-url",
			want:    "",
		},

		// 6. Cas localhost
		{
			name:    "Localhost avec HTTP",
			baseUrl: "http://localhost:8080/app",
			newUrl:  "/api/data",
			want:    "http://localhost:8080/api/data",
		},
		{
			name:    "Localhost avec HTTPS",
			baseUrl: "https://localhost:8443/app/",
			newUrl:  "settings",
			want:    "https://localhost:8443/app/settings",
		},

		// 7. Cas avec fragments à supprimer
		{
			name:    "URL avec fragment à supprimer",
			baseUrl: baseUrl,
			newUrl:  "https://example.com/page#section",
			want:    "https://example.com/page",
		},

		// 8. Cas avec base invalide
		{
			name:    "Base URL invalide",
			baseUrl: "::::invalid-base",
			newUrl:  "/path",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.BuildUrl(tt.baseUrl, tt.newUrl)
			if got != tt.want {
				t.Errorf("\nTest: %s\nBase URL: %s\nNew URL: %s\nGot: %s\nWant: %s",
					tt.name, tt.baseUrl, tt.newUrl, got, tt.want)
			}
		})
	}
}
