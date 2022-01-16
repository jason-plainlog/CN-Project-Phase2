package routes

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"phase2/src/client/api"
	"phase2/src/client/http"
	"strconv"
)

func Chat(req http.HttpRequest, serverChan chan net.Conn) http.HttpResponse {
	tmpl := template.Must(template.ParseFiles("views/chat.html"))

	id := req.Target.Query().Get("id")
	uid, _ := strconv.Atoi(id)

	if req.Method == "POST" {
		_, params, _ := mime.ParseMediaType(req.Headers["Content-Type"])
		mr := multipart.NewReader(bytes.NewReader(req.Data), params["boundary"])
		form, _ := mr.ReadForm(1024576)

		messageType := form.Value["type"][0]

		switch messageType {
		case "text":
			message := form.Value["message"][0]
			api.SendText(uid, message, serverChan)
		case "image":
			fileHeader := form.File["image"][0]
			filename := fileHeader.Filename
			file, _ := fileHeader.Open()
			content, _ := io.ReadAll(file)
			api.SendFile(uid, "image", filename, content, serverChan)
		case "file":
			fileHeader := form.File["file"][0]
			filename := fileHeader.Filename
			file, _ := fileHeader.Open()
			content, _ := io.ReadAll(file)
			api.SendFile(uid, "file", filename, content, serverChan)
		}
	}

	messages, ok := api.GetMessages(uid, serverChan)
	if !ok {
		return http.HttpResponse{
			StatusCode: 404,
			Data:       []byte("404 Not Found"),
		}
	}

	Messages := []struct {
		From, Type, Content string
		IsImage, IsFile     bool
		Link                string
	}{}

	for _, message := range messages {
		Messages = append(Messages, struct {
			From, Type, Content string
			IsImage, IsFile     bool
			Link                string
		}{
			From:    message.From,
			Type:    message.Type,
			Content: message.Content,
			IsImage: message.Type == "image",
			IsFile:  message.Type == "file",
			Link:    fmt.Sprintf("/get?%s=%s", message.Type, message.Content),
		})
	}

	var buffer bytes.Buffer
	tmpl.Execute(&buffer, struct {
		Messages []struct {
			From, Type, Content string
			IsImage, IsFile     bool
			Link                string
		}
	}{Messages: Messages})

	return http.HttpResponse{
		StatusCode: 200,
		Data:       buffer.Bytes(),
	}
}
