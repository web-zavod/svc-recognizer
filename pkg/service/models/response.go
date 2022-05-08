package models

type Category struct {
	ID       string `json:"id"`
	Category string `json:"name"`
}

type Source struct {
	ID    string `json:"id"`
	Query struct {
		Name string `json:"name"`
	} `json:"query"`
}

type SearchResponse struct {
	Took int64
	Hits struct {
		Total struct {
			Value int64
		}
		Hits []*SearchHit `json:"hits"`
	}
}

type SearchHit struct {
	Index  string  `json:"_index"`
	Type   string  `json:"_type"`
	Score  float64 `json:"_score"`
	ID     string  `json:"_id"`
	Source Source  `json:"_source"`
}

type ErrorResponse struct {
	Info   ErrorInfo `json:"error"`
	Status int32     `json:"status"`
}

type ErrorInfo struct {
	RootCause []*ErrorInfo `json:"root_cause"`
	Type      string       `json:"type"`
	Reason    string       `json:"reason"`
}
