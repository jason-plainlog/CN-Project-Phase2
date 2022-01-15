package api

import (
	"encoding/base64"
	"fmt"
	"net"
	"os"
)

func GetMessages(uid int, serverChan chan net.Conn) ([]struct {
	From, Type, Content string
}, bool) {
	results := []struct{ From, Type, Content string }{}
	var reply string
	var entries int

	conn := <-serverChan
	fmt.Fprintln(conn, "chat", uid)
	fmt.Fscanf(conn, "%s", &reply)

	if reply != "ok" {
		serverChan <- conn
		return results, false
	}

	fmt.Fscanf(conn, "%d", &entries)
	for i := 0; i < entries; i++ {
		var messageType, from, content string
		fmt.Fscanf(conn, "%s %s %s", &messageType, &from, &content)

		decodedContent, _ := base64.StdEncoding.DecodeString(content)
		results = append(results, struct {
			From, Type, Content string
		}{
			From:    from,
			Type:    messageType,
			Content: string(decodedContent),
		})

		if messageType != "text" {
			var filename, content string
			fmt.Fscanf(conn, "%s", &filename)

			path := fmt.Sprintf(
				"client_dir/%ss/%s_%s",
				messageType,
				decodedContent,
				filename,
			)

			_, err := os.Stat(path)
			if err != nil {
				fmt.Fprintln(conn, "get")
				fmt.Fscanf(conn, "%s", &content)
				decodedContent, _ := base64.StdEncoding.DecodeString(content)
				os.WriteFile(path, decodedContent, 0644)
			} else {
				fmt.Fprintln(conn, "got")
			}
		}
	}

	serverChan <- conn

	return results, true
}

func SendFile(
	uid int,
	filetype string,
	filename string,
	content []byte,
	serverChan chan net.Conn,
) bool {
	conn := <-serverChan
	var reply string

	encoded := base64.StdEncoding.EncodeToString(content)

	fmt.Fprintln(conn, "send", filetype, uid, filename, encoded)
	fmt.Fscanf(conn, "%s", &reply)

	serverChan <- conn
	return reply == "ok"
}

func SendText(uid int, message string, serverChan chan net.Conn) bool {
	conn := <-serverChan

	message = base64.StdEncoding.EncodeToString([]byte(message))
	var reply string
	fmt.Fprintln(conn, "send text", uid, message)
	fmt.Fscanf(conn, "%s", &reply)

	serverChan <- conn

	return reply == "ok"
}

func AddFriend(username string, serverChan chan net.Conn) bool {
	conn := <-serverChan

	var reply string
	fmt.Fprintln(conn, "add", username)
	fmt.Fscanf(conn, "%s", &reply)

	serverChan <- conn

	return reply == "ok"
}

func DeleteFriend(username string, serverChan chan net.Conn) bool {
	conn := <-serverChan

	var reply string
	fmt.Fprintln(conn, "delete", username)
	fmt.Fscanf(conn, "%s", &reply)

	serverChan <- conn

	return reply == "ok"
}

func GetFriends(serverChan chan net.Conn) []struct {
	Id       int
	Username string
} {
	conn := <-serverChan
	fmt.Fprintln(conn, "list")

	var results []struct {
		Id       int
		Username string
	}

	var entries, id int
	var username string
	fmt.Fscanf(conn, "%d", &entries)
	for i := 0; i < entries; i++ {
		fmt.Fscanf(conn, "%d %s", &id, &username)
		results = append(results, struct {
			Id       int
			Username string
		}{
			Id:       id,
			Username: username,
		})
	}

	serverChan <- conn
	return results
}
