package indexation

import (
	"search_egine/db"
	"testing"
)

func TestGetPages(t *testing.T) {
	storage, err := db.NewStorage("search_engine")
	if err != nil {
		t.Error(err)
	}
	pages := db.GetPages(storage, 10)
	if len(pages) == 0 {
		t.Error("Nous attendions 10 pages")
	}
	if len(pages) > 10 {
		t.Error("Nous attendions 10 pages")
	}
}
