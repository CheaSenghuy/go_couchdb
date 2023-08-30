package models

type Student struct {
	ID    string `json:"_id,omitempty"`
	Rev   string `json:"_rev,omitempty"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Class string `json:"class"`
}
