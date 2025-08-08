package main

import (
	"context"
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
	username := r.FormValue("username")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	fmt.Println("CONNECTED")
	defer conn.Close()

	clients[username].Conn = conn
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

	// Find the call document for the sender
	filter := bson.M{
		"$or": []bson.M{
			{"userA": sender},
			{"userB": sender},
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
