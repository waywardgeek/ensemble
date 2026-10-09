package common

import "encoding/json"

// JSONBounds describes syntax/resource limits, never protocol policy. Zero Nodes
// means no additional total-node bound; Bytes/Depth/Collection must be positive.
type JSONBounds struct {
	Bytes, Depth, Nodes, Collection int
	Scalars                         bool
}
type JSONService interface {
	Ensemble() Ensemble
	Parse([]byte, JSONBounds) (any, error)
	Encode(any, bool, int) ([]byte, error)
	ScalarEscapes([]byte) error
	Number(string) string
	Compare(json.Number, json.Number) int
	Integral(json.Number) bool
}
