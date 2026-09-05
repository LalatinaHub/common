package model

// KeyValue represents a generic key-value configuration entry.
type KeyValue struct {
	ID    int64  `json:"id"`
	Key   string `json:"key"`
	Value any    `json:"value"`
}
