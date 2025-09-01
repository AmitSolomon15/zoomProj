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

var (
	clients                = make(map[string]*Client)
	mutex                  sync.Mutex
	client                 *mongo.Client
	stdin                  io.WriteCloser
	stdout                 io.ReadCloser
	ffmpegOutChan          = make(chan []byte, 1024)
	ebmlHeader             []byte
	headerCaptured         bool
	clientsConnectionExist = make(map[string]bool)
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
	http.HandleFunc("/disconnectWs", disconnectHandler)
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
		"-loglevel", "debug",
		"-fflags", "nobuffer+discardcorrupt",
		"-flags", "low_delay",
		"-probesize", "50M",
		"-analyzeduration", "100M",
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
		"-vsync", "0",
		"-muxdelay", "0",
		"-muxpreload", "0",
		"-f", "mp4",
		"-movflags", "+frag_keyframe+empty_moov+default_base_moof+separate_moof+separate_moof+frag_every_frame",
		"pipe:1",
	)
	stdin, _ = excmd.StdinPipe()
	stdout, _ = excmd.StdoutPipe()
	//fmt.Println("OUT: ", stdout)
	excmd.Stderr = os.Stderr
	excmd.Start()
	reader := bufio.NewReader(stdout)
	go func() {

		fmt.Println("ENTERED FIRST FUNC")
		buf := make([]byte, 1024*64)
		fmt.Println("READ FROM STDOUT")
		for {
			//fmt.Println("READABLE BITS ", reader.Buffered())
			//if stdout != nil && reader.Buffered() > 0 {
			//fmt.Println("ENTERED FIRST FUNC LOOOP")
			//fmt.Println("READABLE BITS ", reader.Buffered())
			//fmt.Println(stdout.Read(buf))
			//fmt.Println("STDOUT")
			n, err := reader.Read(buf)

			//fmt.Println("buffer ", string(buf[:n]))
			if err != nil {
				fmt.Println("ffmpeg stdout error:", err)

				return
			}
			// copy to avoid re-use of buf
			data := make([]byte, n)
			copy(data, buf[:n])
			//fmt.Println("buffer2 ", string(data))
			ffmpegOutChan <- data
			//}

		}
	}()
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	//countMP4 = 0
	username, conn := connectWS(w, r)
	fmt.Printf("User %s connected\n", username)

	go forwordToReciver(username)

	// Listen for messages
	found := false

	for {
		user2, err := findReciever(username)
		if err != nil {
			fmt.Println("ERROR ", err)
			fmt.Println("BOOLS ", len(clientsConnectionExist))
			fmt.Println("CLIENTS ", len(clients))
			fmt.Println("found ", found)
			if len(clientsConnectionExist) < len(clients) {
				fmt.Println("RETURN")
				return
			}
			continue
		} else {
			clientsConnectionExist[user2] = true
		}

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
			found = false
			fmt.Println("MP4")
			mutex.Lock()
			//fmt.Println("mp4 ", string(msg))
			conn.WriteMessage(websocket.BinaryMessage, msg)
			mutex.Unlock()
			continue
		} else {
			fmt.Println("ISWEB: ", isWebM(msg))
			fmt.Println("FOUNd: ", found)
			if found || isWebM(msg) {
				fmt.Println("SENDING")
				handleIncoming(msg)
				found = true
			}

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

func forwordToReciver(sender string) {
	fmt.Println("forwordToReciver")
	receiver, err := findReciever(sender)
	if err != nil {
		fmt.Println("ERROR ", err)
		return
	}

	receiverConn := clients[receiver].Conn
	go func() {
		//var sentChunk []byte
		for chunk := range ffmpegOutChan {
			//fmt.Println("CHUNK: ", chunk)
			//if countMP4 == 2 {
			//countMP4 = 0
			mutex.Lock()
			err := receiverConn.WriteMessage(websocket.BinaryMessage, chunk)
			mutex.Unlock()
			if err != nil {
				fmt.Println("write error to receiver:", err)
				return
			}
			//} else {
			//countMP4++
			//sentChunk = append(sentChunk, chunk...)
			//}
		}
	}()
}

func handleIncoming(data []byte) {
	if !headerCaptured {
		idx := bytes.Index(data, []byte{0x1F, 0x43, 0xB6, 0x75})
		if idx > 0 {
			ebmlHeader = append([]byte{}, data[:idx]...)
			headerCaptured = true
			//fixedData = data
			mutex.Lock()
			stdin.Write(ebmlHeader)
			mutex.Unlock()
			data = data[idx:]
			mutex.Lock()
			stdin.Write(data)
			mutex.Unlock()
		}
	} else {
		mutex.Lock()
		stdin.Write(data)
		mutex.Unlock()
	}
	//fmt.Println(data)

	fmt.Println("WRITTEN TO STDIN")

}
func disconnectHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if clients[username].Conn != nil {
		delete(clients, username)
		delete(clientsConnectionExist, username)
	}
}
