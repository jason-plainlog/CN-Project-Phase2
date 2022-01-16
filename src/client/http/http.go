package http

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type HttpRequest struct {
	Method  string
	Target  *url.URL
	Headers map[string]string
	Data    []byte
}

type HttpResponse struct {
	StatusCode int
	Headers    map[string]string
	Data       []byte
}

func SendResponse(res HttpResponse, conn net.Conn) {
	fmt.Fprintln(conn, "HTTP/1.1", res.StatusCode)
	for key, value := range res.Headers {
		fmt.Fprintf(conn, "%s: %s\n", key, value)
	}
	fmt.Fprintln(conn, "Content-Length:", len(res.Data))
	fmt.Fprintln(conn)
	conn.Write(res.Data)
}

func ParseRequest(conn net.Conn) (HttpRequest, error) {
	var method, target string
	headers := make(map[string]string)

	_, err := fmt.Fscanf(conn, "%s %s HTTP/1.1\n", &method, &target)
	if err != nil {
		return HttpRequest{}, err
	}

	URL, _ := url.Parse(target)

	reader := bufio.NewReader(conn)
	for {
		line, _, _ := reader.ReadLine()
		if len(line) == 0 {
			break
		}
		splitted := strings.SplitN(string(line), ": ", 2)
		headers[splitted[0]] = splitted[1]
	}

	var content []byte
	if value, ok := headers["Content-Length"]; ok {
		length, _ := strconv.Atoi(value)
		content = make([]byte, length)

		io.ReadFull(reader, content)
	}

	return HttpRequest{
		Method:  method,
		Target:  URL,
		Headers: headers,
		Data:    content,
	}, nil
}
