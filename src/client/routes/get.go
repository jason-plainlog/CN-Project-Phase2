package routes

import (
	"net"
	"os"
	"phase2/src/client/http"
	"strings"
)

func Get(req http.HttpRequest, serverChan chan net.Conn) http.HttpResponse {
	image := req.Target.Query().Get("image")
	file := req.Target.Query().Get("file")

	var content []byte

	headers := map[string]string{}

	if image != "" {
		entries, _ := os.ReadDir("client_dir/images/")
		for _, de := range entries {
			if strings.HasPrefix(de.Name(), image) {
				content, _ = os.ReadFile("client_dir/images/" + de.Name())
				break
			}
		}
	} else if file != "" {
		entries, _ := os.ReadDir("client_dir/files/")
		for _, de := range entries {
			if strings.HasPrefix(de.Name(), file) {
				content, _ = os.ReadFile("client_dir/files/" + de.Name())
				headers["Content-Disposition"] = "attachment; filename=" + de.Name()
				break
			}
		}
	}

	return http.HttpResponse{
		StatusCode: 200,
		Data:       content,
		Headers:    headers,
	}
}
