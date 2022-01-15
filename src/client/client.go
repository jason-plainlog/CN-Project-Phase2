package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
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
		conn := <-serverChan
		fmt.Fprintf(conn, "chat %d\n", id)

		var reply string
		fmt.Fscanf(conn, "%s", &reply)

		if reply != "ok" {
			serverChan <- conn
			break
		}

		var entries int
		fmt.Fscanf(conn, "%d", &entries)

		clearScreen()
		if information != "" {
			fmt.Println(information + "\n")
		}
		fmt.Printf("Message History (%d):\n", entries)
		for i := 0; i < entries; i++ {
			var messageType, from, content string
			fmt.Fscanf(conn, "%s %s %s", &messageType, &from, &content)

			decodedContent, _ := base64.StdEncoding.DecodeString(content)

			if messageType == "image" {
				var filename, imageContent string
				fmt.Fscanf(conn, "%s", &filename)
				_, err := os.Stat("client_dir/images/" + string(decodedContent) + "_" + filename)
				if err != nil {
					fmt.Fprintln(conn, "get")
					fmt.Fscanf(conn, "%s", &imageContent)
					decodedImage, _ := base64.StdEncoding.DecodeString(imageContent)
					os.WriteFile(
						"client_dir/images/"+string(decodedContent)+"_"+filename,
						decodedImage,
						0644,
					)
				} else {
					fmt.Fprintln(conn, "got")
				}
			} else if messageType == "file" {
				var filename, fileContent string
				fmt.Fscanf(conn, "%s", &filename)
				_, err := os.Stat("client_dir/files/" + string(decodedContent) + "_" + filename)
				if err != nil {
					fmt.Fprintln(conn, "get")
					fmt.Fscanf(conn, "%s", &fileContent)
					decodedFile, _ := base64.StdEncoding.DecodeString(fileContent)
					os.WriteFile(
						"client_dir/files/"+string(decodedContent)+"_"+filename,
						decodedFile,
						0644,
					)
				} else {
					fmt.Fprintln(conn, "got")
				}
			}

			fmt.Printf("    %s: %s\n", from, decodedContent)
		}
		serverChan <- conn

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
			message = base64.StdEncoding.EncodeToString([]byte(message))

			conn := <-serverChan
			fmt.Fprintf(conn, "send text %d %s\n", id, message)
			fmt.Fscanf(conn, "%s", &reply)
			serverChan <- conn
			information = "message sent!"
		} else if action == 2 {
			fmt.Print("path to image: ")
			fmt.Scanf("%s", &path)
			filename, content, err := readFile(path)
			if err != nil {
				information = err.Error()
				continue
			}

			encoded := base64.StdEncoding.EncodeToString(content)

			conn := <-serverChan
			fmt.Fprintf(conn, "send image %d %s %s\n", id, filename, encoded)
			fmt.Fscanf(conn, "%s", &reply)
			serverChan <- conn
			information = "image sent!"
		} else if action == 3 {
			fmt.Print("path to file: ")
			fmt.Scanf("%s", &path)
			filename, content, err := readFile(path)
			if err != nil {
				information = err.Error()
				continue
			}
			encoded := base64.StdEncoding.EncodeToString(content)

			conn := <-serverChan
			fmt.Println("!")
			fmt.Fprintf(conn, "send file %d %s %s\n", id, filename, encoded)
			fmt.Fscanf(conn, "%s", &reply)
			serverChan <- conn
			information = "file sent!"
		} else if action == 4 {
			break
		}
	}
}

func promptHandler(serverChan chan net.Conn) {
	var action int

	for {
		fmt.Println("\nHome:")
		fmt.Println("(1) List all friends")
		fmt.Println("(2) Add friend")
		fmt.Println("(3) Delete friend")
		fmt.Println("(4) Choose a chatroom")
		fmt.Print("\n> ")
		fmt.Scanf("%d", &action)

		if action == 1 {
			clearScreen()
			conn := <-serverChan
			fmt.Fprintln(conn, "list")

			var entries, id int
			var username string
			fmt.Fscanf(conn, "%d", &entries)
			fmt.Printf("All Friends (%d):\n", entries)
			for i := 0; i < entries; i++ {
				fmt.Fscanf(conn, "%d %s", &id, &username)
				fmt.Printf("    - %s\n", username)
			}
			serverChan <- conn
		} else if action == 2 {
			var username, reply string
			fmt.Print("username: ")
			fmt.Scanf("%s", &username)

			conn := <-serverChan
			fmt.Fprintln(conn, "add", username)

			fmt.Fscanf(conn, "%s", &reply)
			clearScreen()
			if reply == "ok" {
				fmt.Println("Successfully added " + username + " as friend")
			} else {
				fmt.Println("User " + username + " doesn't exist")
			}

			serverChan <- conn
		} else if action == 3 {
			var username, reply string
			fmt.Print("username: ")
			fmt.Scanf("%s", &username)

			conn := <-serverChan
			fmt.Fprintln(conn, "delete", username)

			fmt.Fscanf(conn, "%s", &reply)
			clearScreen()
			if reply == "ok" {
				fmt.Println("Successfully delete friend " + username)
			} else {
				fmt.Println("Friend " + username + " doen't exist")
			}

			serverChan <- conn
		} else if action == 4 {
			clearScreen()
			conn := <-serverChan
			fmt.Fprintln(conn, "list")

			var entries, id int
			var username string
			fmt.Fscanf(conn, "%d", &entries)
			fmt.Printf("Chatrooms (%d):\n", entries)
			for i := 0; i < entries; i++ {
				fmt.Fscanf(conn, "%d %s", &id, &username)
				fmt.Printf("    [%d] %s\n", id, username)
			}
			serverChan <- conn

			fmt.Print("\nenter chatroom [?]: ")
			fmt.Scanf("%d", &id)

			chatHandler(id, serverChan)
			clearScreen()
		} else {
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
