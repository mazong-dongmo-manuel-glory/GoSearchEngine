package models

type Word struct {
	Value     string
	Count     int64
	TfIdf     int
	Url       string
	Occurence int64
}
