package parser_test

import (
	"search_egine/parser"
	"testing"
)

func TestNewParser(t *testing.T) {
	p := parser.NewParser(`<!DOCTYPE html>
<html>
<head>
<title>Page de test pour parsing Go</title>
</head>
<body>

<h1>Liens de test</h1>

<ul>
  <li><a href="https://www.example.com/page1">Page 1</a></li>
  <li><a href="/page2.html">Page 2 (relative)</a></li>
  <li><a href="https://sub.example.com/test/index.php?id=5&name=test">Sous-domaine avec paramètres</a></li>
  <li><a href="#section3">Section 3 (ancre)</a></li>
  <li><a href="mailto:test@example.com">Contactez-nous</a></li>
</ul>

</body>
</html>`, `https://www.example.com/page1`)

	p.Traverse()
	if len(p.Url) != 4 {
		t.Error("Nous attendions 4 urls ", p.Url)
	}

}
