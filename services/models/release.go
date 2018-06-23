package models

type Release struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	PodLabels []Labels `json:"podLabels"`
}

type Labels map[string]string
