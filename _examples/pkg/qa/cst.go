package qa

import (
	"github.com/go-ego/gse"
)

type QAPair struct {
	Question string
	Answer   string
}

var (
	seg gse.Segmenter
)

type Vector map[string]int

type VectorI map[string]int64

type VectorF map[string]float64
