package domain

type HTTPRequest struct {
	URL         string
	Method      string
	Headers     map[string]string
	Body        []byte
	Cookies     map[string]string
	QueryParams map[string]string
}

type HTTPResponse struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
	BodyStr    string
	URL        string
	Time       float64
}
