package models

type Release struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Gateway Gateway `json:"gateway"`
	Apps    []App   `json:"apps"`
}

type Gateway struct {
	Hosts []string `json:"hosts"`
}

type App struct {
	Hosts  []string `json:"hosts"`
	Labels Labels   `json:"labels"`
}

type Labels map[string]string

// type Release struct {
// 	ID        string   `json:"id"`
// 	Name      string   `json:"name"`
// 	PodLabels []Labels `json:"podLabels"`
// }

// type Labels map[string]string
