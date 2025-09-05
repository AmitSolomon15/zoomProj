package main

import (
	"bufio"
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
type CallSession struct {
	Cmd     *exec.Cmd
	Stdin   io.WriteCloser
	Stdout  io.ReadCloser
	OutChan chan []byte
}

var (
	clients = make(map[string]*Client)
	mutex   sync.Mutex
	client  *mongo.Client
	//stdin                  io.WriteCloser
	//stdout                 io.ReadCloser
	//ffmpegOutChan          = make(chan []byte, 1024)
	//ebmlHeader             []byte
	//headerCaptured         bool
	clientsConnectionExist = make(map[string]bool)
	callSessions           = make(map[string]*CallSession)
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
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/disconnectWs", disconnectHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.ListenAndServe(":"+port, nil)

}
func buildCallID(user1, user2 string) string {
	if user1 < user2 {
		return user1 + "_" + user2
	}
	return user2 + "_" + user1
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
func cmdInit(callID string) (*CallSession, error) {
	cmd := exec.Command("ffmpeg",
		"-loglevel", "debug",
		"-fflags", "nobuffer+discardcorrupt",
		"-flags", "low_delay",
		"-probesize", "50M",
		"-analyzeduration", "100M",
		"-re",
		"-f", "webm", // input format
		"-i", "pipe:0", // read from stdin
		"-ac", "2",
		"-preset", "ultrafast",
		"-c:v", "libx264",
		"-c:a", "aac",
		"-b:a", "128k",
		"-ar", "48000",
		"-profile:v", "baseline",
		"-level", "3.1",
		"-x264-params", "keyint=10:scenecut=0", // shorter GOP -> more keyframes
		"-flush_packets", "1",
		"-avioflags", "direct",
		"-g", "1", // force keyframe interval
		"-frag_duration", "1000000",
		"-vsync", "0",
		"-muxdelay", "0",
		"-muxpreload", "0",
		"-f", "mp4",
		"-movflags", "+frag_keyframe+empty_moov+default_base_moof+separate_moof+separate_moof+frag_every_frame",
		"pipe:1",
	)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr

	session := &CallSession{
		Cmd:     cmd,
		Stdin:   stdin,
		Stdout:  stdout,
		OutChan: make(chan []byte, 100),
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// goroutine to read ffmpeg stdout → OutChan
	go func() {
		reader := bufio.NewReader(stdout)
		buf := make([]byte, 64*1024)
		for {
			n, err := reader.Read(buf)
			if err != nil {
				close(session.OutChan)
				return
			}
			data := make([]byte, n)
			copy(data, buf[:n])
			session.OutChan <- data
		}
	}()

	mutex.Lock()
	callSessions[callID] = session
	mutex.Unlock()

	return session, nil
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	//countMP4 = 0
	username, sUsernam, conn := connectWS(w, r)

	callID := buildCallID(username, sUsernam)

	mutex.Lock()
	session, ok := callSessions[callID]
	mutex.Unlock()

	if !ok {
		session, _ = cmdInit(callID)
	}

	fmt.Printf("User %s connected\n", username)

	//go forwordToReciver(callID, username)
	if clients[sUsernam].Conn != nil {
		go forwordToReciver(callID, sUsernam)

		fmt.Println("AFTER FORWORDING")
		for {
			_, err := findReciever(username)
			if err != nil {
				fmt.Println("ERROR ", err)
				if err.Error() == "found no call" {
					break
				}
				continue
			}

			// Read a message from this client
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Read error:", err)
				break
			}

			// MP4 chunks just echo back to sender
			if isMp4(msg) {
				conn.WriteMessage(websocket.BinaryMessage, msg)
				continue
			}

			// WebM chunks are forwarded to ffmpeg (session stdin)
			if isWebM(msg) {
				session.Stdin.Write(msg)
			}

		}

		// Clean up on disconnect
		fmt.Println("DELETING")
		delete(clients, username)
	}
}

func connectWS(w http.ResponseWriter, r *http.Request) (string, string, *websocket.Conn) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return "", "", nil
	}
	fmt.Println("CONNECTED")

	username := r.URL.Query().Get("username")
	sUsername := r.URL.Query().Get("sUsername")

	clients[username] = &Client{Conn: conn}

	return username, sUsername, conn
}

func isMp4(msg []byte) bool {
	if len(msg) < 12 {
		return false
	}

	// WebM magic header
	if bytes.Equal(msg[:4], []byte{0x1A, 0x45, 0xDF, 0xA3}) {
		return false
	}

	// MP4 usually has "ftyp" at offset 4
	format := string(msg[4:8])
	//fmt.Println("FORMAT: ", format)

	if format == "ftyp" || format == "isom" || format == "moov" || format == "mdat" || format == "moof" || format == "udta" {
		return true
	}

	/*
		if bytes.Contains(msg, []byte("moof")) || bytes.Contains(msg, []byte("ftyp")) || bytes.Contains(msg, []byte("isom")) || bytes.Contains(msg, []byte("moov")) || bytes.Contains(msg, []byte("mdat")) || bytes.Contains(msg, []byte("udta")) {
			return true
		}
		if countMP4 == 1 {
			return true
		}
	*/
	return false
}

func isWebM(msg []byte) bool {
	return len(msg) >= 4 && bytes.Equal(msg[0:4], []byte{0x1A, 0x45, 0xDF, 0xA3})
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
		delete(clientsConnectionExist, receiver)
		return "", errors.New("2nd user not found")
	}

	return receiver, nil
}

func forwordToReciver(callID, receiver string) {
	mutex.Lock()
	session := callSessions[callID]
	mutex.Unlock()

	receiverConn := clients[receiver].Conn
	go func() {
		for chunk := range session.OutChan {
			mutex.Lock()
			err := receiverConn.WriteMessage(websocket.BinaryMessage, chunk)
			mutex.Unlock()
			if err != nil {
				return
			}
		}
	}()
}

func handleIncoming(callID string, data []byte) {
	mutex.Lock()
	session, ok := callSessions[callID]
	mutex.Unlock()
	if !ok {
		return
	}

	session.Stdin.Write(data)
	fmt.Println("WRITTEN TO STDIN")

}
func disconnectHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("users").Collection("usersInCall")
	username := r.URL.Query().Get("username")

	filter := bson.M{
		"$or": []bson.M{
			{"user1": username},
			{"user2": username},
		},
	}

	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		http.Error(w, "Failed to delete session", http.StatusInternalServerError)
		return
	}

	fmt.Println("DISCONNECTING: ", username)
	if clients[username].Conn != nil {
		delete(clients, username)
		delete(clientsConnectionExist, username)
	}
}
