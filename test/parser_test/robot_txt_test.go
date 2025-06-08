package parser

import (
	"search_egine/parser"
	"testing"
)

func TestGetDisallowPath(t *testing.T) {
	rb := &parser.RobotTxt{}
	rb.GetDisallowPath("https://www.camerounweb.com/")
	if rb.DisAllowPath == nil {
		t.Error("Disallow path is nil")
	}

}
func TestCheckIfIsDisAllowPath(t *testing.T) {
	rb := &parser.RobotTxt{}
	rb.GetDisallowPath("https://www.camerounweb.com/")
	if rb.CheckIfIsDisAllowPath("https://www.camerounweb.com/contact") {
		t.Error("Disallow path is not correct")
	}
	if !rb.CheckIfIsDisAllowPath("https://www.camerounweb.com/validate_user.php?url=*") {
		t.Error("Disallow path is not correct")
	}
	
}
