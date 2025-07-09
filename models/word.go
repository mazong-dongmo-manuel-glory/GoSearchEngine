package models

type Word struct {
	Value string
	Count int64
}

type WordPage struct {
	TfIdf int
	Count int
}
