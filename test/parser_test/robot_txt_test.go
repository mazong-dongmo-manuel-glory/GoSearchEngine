package parser

import (
	"search_egine/parser"
	"testing"
)

func TestNewRobotTxt(t *testing.T) {
	rb := parser.RobotTxt{}
	rb.GetDisallowPath("https://www.google.com/")

	if rb.DisallowPath == nil {
		t.Error("Nous attendions un dictionnaire")
	}
	if len(rb.DisallowPath) == 0 {
		t.Error("Nous nous attendions des chemins")
	}
	if len(rb.DisallowPath) != 0 {
		t.Log(rb.DisallowPath)
	}
}

func TestPathIsAllow(t *testing.T) {
	rb := &parser.RobotTxt{}
	rb.GetDisallowPath("https://www.google.com/")
	if !rb.PathIsAllow("https://www.google.com/") {
		t.Error("Nous attendions true")
	}
	if rb.PathIsAllow("https://www.google.com/?hl=*&*&gws_rd=ssl:false") {
		t.Error("Nous attendions false")
	}
}
