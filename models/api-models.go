package models

type User struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Age       int    `json:"age"`
}

type VirtualServiceDTO struct {
	Name        string `json:name`
	Host        string `json.host`
	Subset      string `json.subset`
	ReleaseID   string `json.releaseId`
	ReleaseName string `json.releaseName`
}
