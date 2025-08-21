package main

import (
	"context"
	"encoding/json"
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
	cmd    *exec.Cmd = cmdInit()
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
	//http.HandleFunc("/wsConn", wsConnectHandler)
	//fmt.Println("WebSocket server started on :8080")
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
func cmdInit() *exec.Cmd {
	excmd := exec.Command("ffmpeg",
		"-f", "webm", // webm format
		"-ac", "2", // channels
		"-i", "pipe:0", // read from stdin
		//"-c:v", "h264", //video codec
		"-ar", "48000", // sample rate
		"-f", "mp4", // output format
		"-movflags", "frag_keyframe+empty_moov+default_base_moof", // fragmented MP4 for streaming
		"pipe:1", // write to stdout
	)
	stdin, _ = excmd.StdinPipe()
	stdout, _ = excmd.StdoutPipe()
	return excmd
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	cmd = cmdInit()
	/*
		fmt.Println("PIPE1:")
		stdin, _ = cmd.StdinPipe()
		fmt.Println("PIPE2:")
		stdout, _ = cmd.StdoutPipe()
		fmt.Println("ENTERES WSHNADLER")*/

	username, conn := connectWS(w, r)
	fmt.Printf("User %s connected\n", username)

	// Listen for messages
	for {
		fmt.Println("ENTERED THe LOOP")
		//time.Sleep(time.Second)

		//mutex.Lock()
		msgType, msg, err := conn.ReadMessage()
		//mutex.Unlock()
		//fmt.Println("RECIVED MSG: ", msg)
		if isMp4(msg) {
			json.NewEncoder(w).Encode(msg)
			stdin.Close()
			//stdout.Close()
			stdin, _ = cmd.StdinPipe()
			stdout, _ = cmd.StdoutPipe()
		}

		//fmt.Println("msg recived is: ", string(msg))
		if err != nil {
			fmt.Println("Read error:", err)
			fmt.Println("BREAKING")
			break
		}

		// Handle media forwarding
		fmt.Println("GOING FPRWORD")
		forwardMediaToPeer(username, msgType, msg)

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
	//defer conn.Close()
	fmt.Println("CONNECTED")

	username := r.URL.Query().Get("username")

	clients[username] = &Client{Conn: conn}

	return username, conn
}

func forwardMediaToPeer(sender string, msgType int, msg []byte) {

	fmt.Println("ERRORHA:")
	cmd.Stderr = os.Stderr // so you can debug FFmpeg logs
	cmd.Start()

	mutex.Lock()
	stdin.Write(msg)
	mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	//defer clients[sender].Conn.Close()
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
		User1 string
		User2 string
	}

	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		fmt.Println("No call found for user:", sender)
		return
	}

	fmt.Println("result: ", result)
	fmt.Println("result2: ", result.User1)
	fmt.Println("result3: ", result.User2)

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
		return
	}
	receiverConn := clients[receiver].Conn

	outputMsg := make([]byte, 1024)

	fmt.Println("ABOUT TO READ")

	mutex.Lock()
	len, err := stdout.Read(outputMsg)
	mutex.Unlock()

	fmt.Println("I RAD!")
	if err != nil {
		fmt.Println("Error with ffmpeg: ", err)
		return
	}
	// Forward the media
	//fmt.Println("THE OUTPUT MSG: ", string(outputMsg[:len]))

	//mutex.Lock()
	err = receiverConn.WriteMessage(msgType, outputMsg[:len])
	//mutex.Unlock()

	if err != nil {
		fmt.Println("Error forwarding to", receiver, ":", err)
	}
	fmt.Println("SENT MESSAGE")

}

func isMp4(msg []byte) bool {
	fmt.Println("ENTERED ISMP")
	if len(msg) < 12 {
		return false // too short to be valid
	}
	/*
		chckFTYP := string(msg[4:8])
		chckISOM := string(msg[8:12])
		chkFORMAT := string(msg[16:20])
		validFormats := [...]string{"avc1", "mp41", "iso2", "isom", "mp42"}
		rightFormat := false
		for _, format := range validFormats {
			if chkFORMAT == format {
				rightFormat = true
			}
		}
		fmt.Println("chck1 ", chckFTYP)
		fmt.Println("chck2 ", chckISOM)
		fmt.Println("chck3 ", chkFORMAT)
		if chckFTYP != "ftyp" && chckISOM != "isom" && !rightFormat {
			fmt.Println("NOT MP4")
			return false
		}
		fmt.Println("MP4")
	*/
	header := string(msg[0:4])
	fmt.Println("PRINT HEADER: ", header)
	return header == "1A 45 DF A3"
}
