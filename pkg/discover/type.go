package discover

type Endpoint struct {
    Method      string   `json:"method"`
    Path        string   `json:"path"`
    Description string   `json:"description"`
    Tags        []string `json:"tags,omitempty"`
}

type EndpointList struct {
    Endpoints []Endpoint `json:"endpoints"`
    Source    string     `json:"source"`
    Timestamp string     `json:"timestamp"`
}