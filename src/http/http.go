package http

import "bufio"

type HttpRequest struct {
	Method  int
	Target  string
	Headers map[string]string
	Data    []byte
}

type HttpResponse struct {
	StatusCode int
	Headers    map[string]string
	Data       []byte
}

func ParseRequest(reader *bufio.Reader) {

}
