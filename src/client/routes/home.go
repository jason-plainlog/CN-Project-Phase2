package routes

import (
	"bytes"
	"html/template"
	"net"
	"net/url"
	"phase2/src/client/api"
	"phase2/src/client/http"
)

func Home(req http.HttpRequest, serverChan chan net.Conn) http.HttpResponse {
	tmpl := template.Must(template.ParseFiles("views/home.html"))
	message := ""

	if req.Method == "POST" {
		form, _ := url.ParseQuery(string(req.Data))
		action, username := form.Get("action"), form.Get("username")

		if action == "add" {
			if api.AddFriend(username, serverChan) {
				message = "successfully add " + username + " as friend!"
			} else {
				message = "user " + username + " not exists"
			}
		} else if action == "delete" {
			if api.DeleteFriend(username, serverChan) {
				message = "successfully unfriend user " + username
			} else {
				message = "user " + username + " not exists"
			}
		}
	}

	friends := api.GetFriends(serverChan)

	var buffer bytes.Buffer
	tmpl.Execute(&buffer, struct {
		Friends []struct {
			Id       int
			Username string
		}
		Message string
	}{Friends: friends, Message: message})

	return http.HttpResponse{
		StatusCode: 200,
		Headers:    map[string]string{},
		Data:       buffer.Bytes(),
	}
}
