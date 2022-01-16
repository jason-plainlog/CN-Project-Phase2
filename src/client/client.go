package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"

	"phase2/src/client/api"
	"phase2/src/client/http"
	"phase2/src/client/routes"
)

func getArgs() (string, string) {
	if len(os.Args) != 3 {
		log.Fatalln("usage: " + os.Args[0] + " [server_address] [http_port]")
	}

	serverAddr := os.Args[1]
	httpAddr := "127.0.0.1:" + os.Args[2]
	return serverAddr, httpAddr
}

func auth(conn net.Conn) {
	authed := false

	for !authed {
		var action int
		var username, password string

		fmt.Println("\nAuthentication:")
		fmt.Println("(1) Sign in")
		fmt.Println("(2) Sign up")
		fmt.Print("\n> ")
		fmt.Scanf("%d", &action)

		if action != 1 && action != 2 {
			clearScreen()
			fmt.Println("Invalid action")
			continue
		}

		fmt.Print("username: ")
		fmt.Scanf("%s", &username)
		fmt.Print("password: ")
		fmt.Scanf("%s", &password)

		var reply string
		if action == 1 {
			fmt.Fprintf(conn, "signin %s %s\n", username, password)
			fmt.Fscanf(conn, "%s\n", &reply)

			if reply != "ok" {
				clearScreen()
				fmt.Println("Invalid username or password, please retry.")
			}
		} else if action == 2 {
			fmt.Fprintf(conn, "signup %s %s\n", username, password)
			fmt.Fscanf(conn, "%s\n", &reply)

			if reply != "ok" {
				clearScreen()
				fmt.Println("Username used, please try another.")
			}
		}

		if reply == "ok" {
			authed = true
		}
	}

	clearScreen()
	fmt.Println("Successfully Authenticated.")
}

func clientHandler(conn net.Conn, serverChan chan net.Conn) {
	defer conn.Close()

	get := regexp.MustCompile("^/get")
	chat := regexp.MustCompile("^/chat")
	home := regexp.MustCompile("^/")

	for {
		req, err := http.ParseRequest(conn)
		if err != nil {
			break
		}

		var response http.HttpResponse

		if chat.MatchString(req.Target.Path) {
			response = routes.Chat(req, serverChan)
		} else if get.Match([]byte(req.Target.Path)) {
			response = routes.Get(req, serverChan)
		} else if home.MatchString(req.Target.Path) {
			response = routes.Home(req, serverChan)
		}

		http.SendResponse(response, conn)
	}
}

func readFile(path string) (string, []byte, error) {
	filenames := strings.Split(path, "/")
	filename := filenames[len(filenames)-1]

	content, err := os.ReadFile(path)
	return filename, content, err
}

func chatHandler(id int, serverChan chan net.Conn) {
	information := ""
	for {
		messages, ok := api.GetMessages(id, serverChan)

		if !ok {
			break
		}

		clearScreen()
		if information != "" {
			fmt.Println(information + "\n")
		}
		fmt.Printf("Message History (%d):\n", len(messages))
		for _, message := range messages {
			fmt.Printf("    %s: %s\n", message.From, message.Content)
		}

		fmt.Println("\nAction:")
		fmt.Println("(1) Send Message")
		fmt.Println("(2) Send Image")
		fmt.Println("(3) Send File")
		fmt.Println("(4) Exit")
		fmt.Print("\n> ")

		var action int
		fmt.Scanf("%d", &action)

		var path string
		if action == 1 {
			fmt.Print("message: ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			message := scanner.Text()

			if api.SendText(id, message, serverChan) {
				information = "message sent!"
			} else {
				information = "send message failed!"
			}
		} else if action == 2 {
			fmt.Print("path to image: ")
			fmt.Scanf("%s", &path)
			filename, content, err := readFile(path)
			if err != nil {
				information = err.Error()
				continue
			}

			if api.SendFile(id, "image", filename, content, serverChan) {
				information = "image sent!"
			} else {
				information = "send message failed!"
			}
		} else if action == 3 {
			fmt.Print("path to file: ")
			fmt.Scanf("%s", &path)
			filename, content, err := readFile(path)
			if err != nil {
				information = err.Error()
				continue
			}

			if api.SendFile(id, "file", filename, content, serverChan) {
				information = "file sent!"
			} else {
				information = "send file failed!"
			}
		} else if action == 4 {
			break
		}
	}
}

func promptHandler(serverChan chan net.Conn) {
	var action, id int
	var username string

	for {
		action = 0

		fmt.Println("\nHome:")
		fmt.Println("(1) List all friends")
		fmt.Println("(2) Add friend")
		fmt.Println("(3) Delete friend")
		fmt.Println("(4) Choose a chatroom")
		fmt.Print("\n> ")
		fmt.Scanf("%d", &action)

		if action == 1 {
			clearScreen()

			friends := api.GetFriends(serverChan)
			fmt.Printf("All Friends (%d):\n", len(friends))
			for _, friend := range friends {
				fmt.Printf("    - %s\n", friend.Username)
			}
		} else if action == 2 {
			fmt.Print("username: ")
			fmt.Scanf("%s", &username)

			clearScreen()
			if api.AddFriend(username, serverChan) {
				fmt.Println("Successfully added " + username + " as friend")
			} else {
				fmt.Println("User " + username + " doesn't exist")
			}
		} else if action == 3 {
			fmt.Print("username: ")
			fmt.Scanf("%s", &username)

			clearScreen()
			if api.DeleteFriend(username, serverChan) {
				fmt.Println("Successfully delete friend " + username)
			} else {
				fmt.Println("Friend " + username + " doen't exist")
			}
		} else if action == 4 {
			friends := api.GetFriends(serverChan)

			clearScreen()
			fmt.Printf("Chatrooms (%d):\n", len(friends))
			for _, friend := range friends {
				fmt.Printf("    [%d] %s\n", friend.Id, friend.Username)
			}

			fmt.Print("\nenter chatroom [?]: ")
			fmt.Scanf("%d", &id)

			chatHandler(id, serverChan)
			clearScreen()
		} else {
			clearScreen()
			fmt.Println("Invalid Action")
		}
	}
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func main() {
	os.MkdirAll("client_dir/images", 0755)
	os.MkdirAll("client_dir/files", 0755)

	serverAddr, httpAddr := getArgs()

	serverConn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalln(err)
	}

	auth(serverConn)

	ln, err := net.Listen("tcp", httpAddr)
	if err != nil {
		log.Fatalln(err)
	}

	serverChan := make(chan net.Conn, 1)
	serverChan <- serverConn

	go promptHandler(serverChan)

	for {
		conn, err := ln.Accept()
		if err == nil {
			go clientHandler(conn, serverChan)
		}
	}
}
