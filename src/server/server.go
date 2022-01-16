package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"phase2/src/server/models"

	"github.com/jameycribbs/hare"
	"github.com/jameycribbs/hare/datastores/disk"
	"golang.org/x/crypto/bcrypt"
)

func dbInit() (*disk.Disk, *hare.Database) {
	os.MkdirAll("./server_dir/data", 0755)

	ds, err := disk.New("./server_dir/data", ".json")
	if err != nil {
		panic(err)
	}

	db, err := hare.New(ds)
	if err != nil {
		panic(err)
	}

	db.CreateTable("users")
	db.CreateTable("messages")
	db.CreateTable("files")

	return ds, db
}

func auth(conn net.Conn, db *hare.Database) models.User {
	var user models.User

	authed := false
	for !authed {
		var action, username, password string
		fmt.Fscanf(conn, "%s %s %s\n", &action, &username, &password)

		results, _ := models.QueryUsers(db, func(u models.User) bool {
			return u.Username == username
		}, 1)

		if action == "signin" {
			if len(results) != 1 {
				fmt.Fprintln(conn, "no")
				continue
			}

			user = results[0]
			if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
				fmt.Fprintln(conn, "no")
				continue
			}

			fmt.Fprintln(conn, "ok")
			authed = true
		} else if action == "signup" {
			if len(results) != 0 {
				fmt.Fprintln(conn, "no")
				continue
			}

			passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
			user = models.User{
				Username:     username,
				PasswordHash: string(passwordHash),
				Friends:      map[int]bool{},
			}

			db.Insert("users", &user)
			fmt.Fprintln(conn, "ok")
			authed = true
		}
	}

	return user
}

func handler(conn net.Conn, db *hare.Database) {
	defer conn.Close()

	user := auth(conn, db)

	var command string
	for {
		db.Find("users", user.ID, &user)

		_, err := fmt.Fscanf(conn, "%s", &command)
		if err != nil {
			return
		}

		if command == "list" {
			results, _ := models.QueryUsers(db, func(u models.User) bool {
				isFriend, ok := u.Friends[user.ID]
				return ok && isFriend
			}, 0)

			fmt.Fprintln(conn, len(results))
			for _, u := range results {
				fmt.Fprintln(conn, u.ID, u.Username)
			}
		} else if command == "add" {
			var username string
			fmt.Fscanf(conn, "%s", &username)

			results, _ := models.QueryUsers(db, func(u models.User) bool {
				return u.Username == username
			}, 1)

			if len(results) != 1 {
				fmt.Fprintln(conn, "no")
			} else {
				user2 := results[0]

				if isFriend, ok := user.Friends[user2.ID]; !ok || !isFriend {
					user.Friends[user2.ID] = true
					user2.Friends[user.ID] = true
					db.Update("users", &user)
					db.Update("users", &user2)
				}

				fmt.Fprintln(conn, "ok")
			}
		} else if command == "delete" {
			var username string
			fmt.Fscanf(conn, "%s", &username)

			results, _ := models.QueryUsers(db, func(u models.User) bool {
				return u.Username == username
			}, 1)

			if len(results) != 1 {
				fmt.Fprintln(conn, "no")
			} else {
				user2 := results[0]

				if isFriend, ok := user.Friends[user2.ID]; ok && isFriend {
					user.Friends[user2.ID] = false
					user2.Friends[user.ID] = false
					db.Update("users", &user)
					db.Update("users", &user2)
					fmt.Fprintln(conn, "ok")
				} else {
					fmt.Fprintln(conn, "no")
				}
			}
		} else if command == "chat" {
			var id int
			var user2 models.User
			fmt.Fscanf(conn, "%d", &id)
			err := db.Find("users", id, &user2)

			var isFriend bool
			{
				value, ok := user.Friends[user2.ID]
				isFriend = ok && value
			}
			if err != nil || !isFriend {
				fmt.Fprintln(conn, "no")
			} else {
				fmt.Fprintln(conn, "ok")

				results, _ := models.QueryMessages(db, func(m models.Message) bool {
					fromUser := (m.From == user.Username && m.To == user2.Username)
					fromUser2 := (m.From == user2.Username && m.To == user.Username)

					return fromUser || fromUser2
				}, 0)

				fmt.Fprintln(conn, len(results))
				for _, m := range results {
					var content, reply string
					if m.Type == "text" {
						content = base64.StdEncoding.EncodeToString(m.Content)
						fmt.Fprintln(conn, m.Type, m.From, content)
					} else if m.Type == "image" {
						content = base64.StdEncoding.EncodeToString(
							[]byte(fmt.Sprintf("[img_%d]", m.ID)),
						)
						fileContent := base64.StdEncoding.EncodeToString(
							m.Content,
						)
						fmt.Fprintln(conn, m.Type, m.From, content, m.Filename)
						fmt.Fscanf(conn, "%s", &reply)
						if reply != "got" {
							fmt.Fprintln(conn, fileContent)
						}
					} else if m.Type == "file" {
						content = base64.StdEncoding.EncodeToString(
							[]byte(fmt.Sprintf("[file_%d]", m.ID)),
						)
						fileContent := base64.StdEncoding.EncodeToString(
							m.Content,
						)
						fmt.Fprintln(conn, m.Type, m.From, content, m.Filename)
						fmt.Fscanf(conn, "%s", &reply)
						if reply != "got" {
							fmt.Fprintln(conn, fileContent)
						}
					}
				}
			}
		} else if command == "send" {
			var id int
			var messageType, filename, content string
			fmt.Fscanf(conn, "%s %d", &messageType, &id)

			results, _ := models.QueryUsers(db, func(u models.User) bool {
				return u.ID == id
			}, 0)

			if len(results) != 1 {
				fmt.Fprintln(conn, "no")
			} else if messageType == "text" {
				fmt.Fscanf(conn, "%s", &content)
				user2 := results[0]
				message, _ := base64.StdEncoding.DecodeString(content)

				db.Insert("messages", &models.Message{
					From:      user.Username,
					To:        user2.Username,
					Content:   message,
					Type:      "text",
					Timestamp: time.Now(),
				})
			} else if messageType == "image" {
				fmt.Fscanf(conn, "%s %s", &filename, &content)
				imageContent, _ := base64.StdEncoding.DecodeString(content)

				db.Insert("messages", &models.Message{
					From:      user.Username,
					To:        results[0].Username,
					Type:      "image",
					Filename:  filename,
					Content:   imageContent,
					Timestamp: time.Now(),
				})
			} else if messageType == "file" {
				fmt.Fscanf(conn, "%s %s", &filename, &content)
				fileContent, _ := base64.StdEncoding.DecodeString(content)

				db.Insert("messages", &models.Message{
					From:      user.Username,
					To:        results[0].Username,
					Type:      "file",
					Filename:  filename,
					Content:   fileContent,
					Timestamp: time.Now(),
				})
			}

			fmt.Fprintln(conn, "ok")
		}
	}
}

func getArgs() string {
	if len(os.Args) != 2 {
		log.Fatalln("usage: " + os.Args[0] + " [port]")
	}
	return ":" + os.Args[1]
}

func main() {
	port := getArgs()

	ds, db := dbInit()
	defer ds.Close()
	defer db.Close()

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalln(err)
	}
	for {
		conn, err := listener.Accept()
		if err == nil {
			go handler(conn, db)
		}
	}
}
