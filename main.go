package main

import (
	"log"
	"net/http"
	"serCoba/db"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {  
	conn     *websocket.Conn  
	room     string  
	username string  
}  
  
type Room struct {  
	clients map[*Client]bool  
	mu      sync.Mutex  
}  
  

var rooms = make(map[string]*Room)  
  

func handleConnection(w http.ResponseWriter, r *http.Request) {  
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)  
	if err != nil {  
		log.Println("Error during connection upgrade:", err)  
		return  
	}  
	defer conn.Close()  
  
	var username, room string  
	var action string 
  
	var msg map[string]string  
	if err := conn.ReadJSON(&msg); err != nil {  
		log.Println("Error reading message:", err)  
		return  
	}  
  
	action = msg["action"]  
	username = msg["username"]  
	password := msg["password"]  
	room = msg["room"]  

	log.Println("action :", action)
  
	if action == "login" {  
		valid, err := db.VerifyUser(username, password)  
		if err != nil {  
			conn.WriteJSON(map[string]string{"error": "Error verifying user"})  
			return  
		}  
		if !valid {  
			conn.WriteJSON(map[string]string{"error": "Invalid username or password"})  
			return  
		}  
		conn.WriteJSON(map[string]string{"message": "Login successful"})  
	} else if action == "register" {  
		exists, err := db.VerifyUser(username, password)  
		if err != nil {  
			conn.WriteJSON(map[string]string{"error": "Error checking username"})  
			return  
		}  
		if exists {  
			conn.WriteJSON(map[string]string{"error": "Username already exists"})  
			return  
		}  
  
		err = db.RegisterUser(username, password)  
		if err != nil {  
			conn.WriteJSON(map[string]string{"error": "Error registering user"})  
			return  
		}  
		conn.WriteJSON(map[string]string{"message": "Registration successful"})  
	} else {  
		conn.WriteJSON(map[string]string{"error": "Invalid action"})  
		return  
	}  

	 
	for {  
		var msg map[string]string
		if err := conn.ReadJSON(&msg); err != nil {  
			log.Println("Error reading message:", err)  
			break  
		}  
		msg["username"] = username
		log.Println("username for :", msg["username"])
		log.Println("room for :", msg["room"])
		log.Println("message for :", msg["message"])
		log.Println("action for :", msg["action"])
		log.Println("--------------------------------")
		if msg["action"] == "join" {
			// add client to room
			room = msg["room"]
			client := &Client{conn: conn, room: room, username: msg["username"] }
			log.Println("client :", client)  
			addClientToRoom(client)  
			conn.WriteJSON(map[string]string{"message": "Joined room: " + room})  
			log.Printf("%s joined room: %s", username, room)  
		} else if msg["action"] == "dm" {  
			// DM handler
			log.Println("dm :", msg["action"])
			recipient := msg["to"]  
			message := msg["message"]  
			log.Printf("DM from %s to %s: %s", username, recipient, message)  

		}

		var tipe string
		var receiver string
		if msg["action"] == "dm" {
			tipe = "dm"
			receiver = msg["to"]
			db.SaveMessage(username, tipe, msg["message"], receiver )
		} else {
			if (msg["message"] != "") {
				tipe = "room"
				receiver = "room"
				db.SaveMessage(username, tipe, msg["message"], receiver )
			}
		}
		
		broadcastMessage(room, msg)  
	}  
  
	client := &Client{conn: conn, room: room, username: msg["username"] } 
	removeClientFromRoom(client)  
	log.Printf("%s left room: %s", username, room)  
}  
  
func addClientToRoom(client *Client) {  
	room := rooms[client.room]  
	if room == nil {  
		room = &Room{clients: make(map[*Client]bool)}  
		rooms[client.room] = room  
	}  
	room.mu.Lock()  
	room.clients[client] = true  
	room.mu.Unlock()  
}  
  
func removeClientFromRoom(client *Client) {  
	room := rooms[client.room]  
	if room != nil {  
		room.mu.Lock()  
		delete(room.clients, client)  
		room.mu.Unlock()  
	}  
}  
  

func broadcastMessage(roomName string, msg map[string]string) {  
	room := rooms[roomName]  
	if room != nil {  
		room.mu.Lock()  
		defer room.mu.Unlock()  
		for client := range room.clients {  
			if err := client.conn.WriteJSON(msg); err != nil {  
				log.Println("Error sending message:", err)  
				client.conn.Close()  
				delete(room.clients, client)  
			}  
		}  
	}  
}  
  
func main() {  
	db.InitDB("root:sayakeren@tcp(localhost:3306)/goWebSocket")  
	http.HandleFunc("/ws", handleConnection)  
	log.Println("Server started on :8080")  
	if err := http.ListenAndServe(":8080", nil); err != nil {  
		log.Fatal("ListenAndServe:", err)  
	}  
}  
