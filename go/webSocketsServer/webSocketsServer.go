package main

import (
	"bytes"
	"context"
	"errors"

	//"encoding/json"
	"io"
	"sync"

	//"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	//"golang.org/x/crypto/openpgp/errors"
)

type Client struct {
	Conn     *websocket.Conn
	Username string
}

var (
	clients = make(map[string]*Client)
	mutex   sync.Mutex
	//clientsConnected = make(map[string]bool)
	client *mongo.Client
	stdin  io.WriteCloser
	stdout io.ReadCloser
	//cmd    *exec.Cmd = cmdInit()
	ffmpegOutChan = make(chan []byte, 1024)
)

// Upgrader is used to upgrade HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	fmt.Println("ENTERED MAIN")

	connectMongo()
	cmdInit()
	http.HandleFunc("/ws", wsHandler)
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
func cmdInit() {
	excmd := exec.Command("ffmpeg",
		"-fflags", "+discardcorrupt",
		"-f", "webm", // webm format
		"-i", "pipe:0", // read from stdin
		"-ac", "2", // channels
		"-preset", "ultrafast", // (important for real-time)
		"-c:v", "libx264", // transcode VP8 → H.264
		"-c:a", "aac", // transcode Opus → AAC
		"-b:a", "128k", // audio bitrate
		"-ar", "48000", // sample rate
		"-profile:v", "baseline",
		"-level", "3.1",
		"-x264-params", "keyint=30:scenecut=0",
		"-f", "mp4", // output format
		"-movflags", "+frag_keyframe+empty_moov+default_base_moof", // fragmented MP4 for streaming
		"pipe:1", // write to stdout
	)
	stdin, _ = excmd.StdinPipe()
	stdout, _ = excmd.StdoutPipe()
	excmd.Stderr = os.Stderr
	excmd.Start()

	go func() {
		fmt.Println("ENTERED FIRST FUNC")
		buf := make([]byte, 1024)
		for {
			fmt.Println("ENTERED FIRST FUNC LOOOP")
			n, err := stdout.Read(buf)
			fmt.Println("buffer ", string(buf))
			if err != nil {
				fmt.Println("ffmpeg stdout error:", err)

				return
			}
			// copy to avoid re-use of buf
			data := make([]byte, n)
			copy(data, buf[:n])
			fmt.Println("buffer ", string(data))
			ffmpegOutChan <- data

		}
	}()
}

func wsHandler(w http.ResponseWriter, r *http.Request) {

	username, conn := connectWS(w, r)
	fmt.Printf("User %s connected\n", username)

	forwordToReciver(username)

	// Listen for messages
	for {
		fmt.Println("ENTERED THe LOOP")

		mutex.Lock()
		_, msg, err := conn.ReadMessage()
		mutex.Unlock()

		//fmt.Println("msgType is: ", msgType)

		if err != nil {
			fmt.Println("Read error:", err)
			fmt.Println("BREAKING")
			break
		}
		if isMp4(msg) {
			fmt.Println("MP4")
			mutex.Lock()
			//fmt.Println(string(msg))
			conn.WriteMessage(websocket.BinaryMessage, msg)
			mutex.Unlock()
			continue
		}
		fmt.Println("NOt MP4")

		// Handle media forwarding
		fmt.Println("GOING FPRWORD")
		mutex.Lock()
		_, err = stdin.Write(msg)
		mutex.Unlock()
		if err != nil {
			fmt.Println("Error writing to ffmpeg stdin:", err)
			break
		}
	}

	// Clean up on disconnect
	fmt.Println("DELETING")
	delete(clients, username)
}

func connectWS(w http.ResponseWriter, r *http.Request) (string, *websocket.Conn) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return "", nil
	}
	fmt.Println("CONNECTED")

	username := r.URL.Query().Get("username")

	clients[username] = &Client{Conn: conn}

	return username, conn
}

/*
func forwardMediaToPeer(sender string, msg []byte) {

	fmt.Println("SENDING STDIN")
	mutex.Lock()
	n, _ := stdin.Write(msg)
	mutex.Unlock()
	fmt.Println("WRITTEN ", n)


	receiver, err := findReciever(sender)
	if err != nil {
		fmt.Println("ERROR ", err)
		return
	}

	receiverConn := clients[receiver].Conn


	fmt.Println("ABOUT TO READ")
	go func() {
		fmt.Println("ENTER FUNC 2")
		for chunk := range ffmpegOutChan {
			fmt.Println("ENTER FUNC 2 LOOOP")
			fmt.Println("CHUNK: ", chunk)
			mutex.Lock()
			err := receiverConn.WriteMessage(websocket.BinaryMessage, chunk)
			mutex.Unlock()
			if err != nil {
				fmt.Println("write error to receiver:", err)
				return
			}

			fmt.Println("I RAD!")

			fmt.Println("SENT MESSAGE")
		}
	}()

}*/

func isMp4(msg []byte) bool {
	fmt.Println("ENTERED ISMP")
	if len(msg) < 12 {
		return false // too short to be valid
	}

	header := msg[0:4]
	invalidHeader := []byte{0x1A, 0x45, 0xDF, 0xA3}
	fmt.Println("PRINT HEADER: ", header)
	return !(bytes.Equal(header, invalidHeader) || header[0] == invalidHeader[3] || header[0] >= 0x5A)

}

func findReciever(sender string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("users").Collection("usersInCall")

	//fmt.Println(sender)

	// Find the call document for the sender
	filter := bson.M{
		"$or": []bson.M{
			{"user1": sender},
			{"user2": sender},
		},
	}

	var result struct {
		User1 string
		User2 string
	}

	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		fmt.Println("No call found for user:", sender)
		return "", errors.New("found no call")
	}

	// Determine the receiver
	var receiver string
	if result.User1 == sender {
		receiver = result.User2
	} else {
		receiver = result.User1
	}

	// Check if receiver is connected

	fmt.Println("user ", receiver, " connection: ", clients[receiver])
	if clients[receiver] == nil {
		fmt.Println("Receiver not connected:", receiver)
		return "", errors.New("2nd user not found")
	}
	return receiver, nil
}

func forwordToReciver(sender string) {
	receiver, err := findReciever(sender)
	if err != nil {
		fmt.Println("ERROR ", err)
		return
	}
	receiverConn := clients[receiver].Conn
	go func() {

		for chunk := range ffmpegOutChan {
			fmt.Println("CHUNK: ", chunk)
			mutex.Lock()
			err := receiverConn.WriteMessage(websocket.BinaryMessage, chunk)
			mutex.Unlock()
			if err != nil {
				fmt.Println("write error to receiver:", err)
				return
			}
		}
	}()
}
