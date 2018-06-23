package models

type TrafficSegment struct {
	ID    string     `json:"id"`
	Name  string     `json:"name"`
	Match *HttpMatch `json:"match"`
}

type HttpMatch struct {
	Headers map[string]*StringMatch `json:"headers,omitempty"`
}

type StringMatch struct {
	Regex string `json:"regex"`
}
