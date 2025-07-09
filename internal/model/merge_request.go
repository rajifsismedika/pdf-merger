package model

type MergeRequest struct {
	Name string   `json:"name"`
	URLs []string `json:"urls"`
}

type MergeWrapper struct {
	Status string       `json:"status"`
	Data   MergeRequest `json:"data"`
}
