package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	Conn     *websocket.Conn
	Username string
}

var (
	clients = make(map[string]*Client)
	client  *mongo.Client
)

// Upgrader is used to upgrade HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	connectMongo()
	http.HandleFunc("/ws", wsHandler)
	fmt.Println("WebSocket server started on :8080")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.ListenAndServe(":"+port, nil)

}

func connectMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://amitsol462:AmitS210706@cluster0.jbild9v.mongodb.net/"))
	if err != nil {
		panic(err)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ENTERES WSHNADLER")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	fmt.Println("CONNECTED")
	defer conn.Close()
	var username string

	// First message should be JSON with username
	_, msg, err := conn.ReadMessage()
	if err != nil {
		fmt.Println("Error reading username:", err)
		return
	}

	var initData struct {
		Type     string
		Username string
	}
	if err := json.Unmarshal(msg, &initData); err != nil {
		fmt.Println("JSON parse error:", err)
		return
	}
	username = initData.Username
	fmt.Printf("User %s connected\n", username)

	clients[username] = &Client{Conn: conn}

	fmt.Printf("User %s connected\n", username)

	// Listen for messages
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read error:", err)
			break
		}

		// Handle media forwarding
		forwardMediaToPeer(username, msgType, msg)
	}

	// Clean up on disconnect
	delete(clients, username)
}

func forwardMediaToPeer(sender string, msgType int, msg []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("users").Collection("usersInCall")

	fmt.Println(sender)

	// Find the call document for the sender
	filter := bson.M{
		"$or": []bson.M{
			{"user1": sender},
			{"user2": sender},
		},
	}

	var result struct {
		UserA string
		UserB string
	}

	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		fmt.Println("No call found for user:", sender)
		return
	}
	fmt.Println("result: ", result)
	fmt.Println("result2: ", result.UserA)
	fmt.Println("result3: ", result.UserB)

	// Determine the receiver
	var receiver string
	if result.UserA == sender {
		receiver = result.UserB
	} else {
		receiver = result.UserA
	}

	// Check if receiver is connected
	receiverConn, ok := clients[receiver]
	if !ok {
		fmt.Println("Receiver not connected:", receiver)
		return
	}

	// Forward the media
	err = receiverConn.Conn.WriteMessage(msgType, msg)
	if err != nil {
		fmt.Println("Error forwarding to", receiver, ":", err)
	}
}
