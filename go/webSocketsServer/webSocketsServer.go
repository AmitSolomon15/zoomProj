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
	clients = make(map[string]*Client)
	mutex   sync.Mutex
	//clientsConnected = make(map[string]bool)
	client *mongo.Client
	stdin  io.WriteCloser
	stdout io.ReadCloser
	//cmd    *exec.Cmd = cmdInit()
	ffmpegOutChan = make(chan []byte, 1024)
	/*
		clusterBuf     bytes.Buffer
		insideCluster  bool
		clusterSize    int
	*/
	ebmlHeader     []byte
	headerCaptured bool
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
		"-g", "10", // force keyframe interval
		"-vsync", "0",
		"-f", "mp4",
		"-movflags", "+frag_keyframe+empty_moov+default_base_moof+delay_moov",
		"pipe:1",
	)
	stdin, _ = excmd.StdinPipe()
	stdout, _ = excmd.StdoutPipe()
	//fmt.Println("OUT: ", stdout)
	excmd.Stderr = os.Stderr
	excmd.Start()

	go func() {
		reader := bufio.NewReader(stdout)
		fmt.Println("ENTERED FIRST FUNC")
		buf := make([]byte, 1024*64)
		fmt.Println("READ FROM STDOUT")
		fmt.Println("READABLE BITS ", reader.Buffered())
		for {
			fmt.Println("ENTERED FIRST FUNC LOOOP")

			if stdout != nil {
				//fmt.Println(stdout.Read(buf))
				//fmt.Println("STDOUT")
				n, err := reader.Read(buf)

				fmt.Println("buffer ", string(buf))
				if err != nil {
					fmt.Println("ffmpeg stdout error:", err)

					return
				}
				// copy to avoid re-use of buf
				data := make([]byte, n)
				copy(data, buf[:n])
				//fmt.Println("buffer2 ", string(data))
				ffmpegOutChan <- data
			}
		}
	}()
}

func wsHandler(w http.ResponseWriter, r *http.Request) {

	username, conn := connectWS(w, r)
	fmt.Printf("User %s connected\n", username)

	forwordToReciver(username)

	// Listen for messages
	found := false
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
			found = false
			fmt.Println("MP4")
			mutex.Lock()
			fmt.Println("mp4 ", string(msg))
			conn.WriteMessage(websocket.BinaryMessage, msg)
			mutex.Unlock()
			continue
		} else {
			/*
				//fmt.Println(isMp4(msg))
				// Handle media forwarding
				fmt.Println("GOING FPRWORD")
				mutex.Lock()
				fmt.Println("THE DATA SENT TO FFMPEG: ", msg)
				numOfBytes, err := stdin.Write(msg)
				fmt.Println("size of msg: ", numOfBytes)
				mutex.Unlock()
				//fmt.Println("READ")
				if err != nil {
					fmt.Println("Error writing to ffmpeg stdin:", err)
					break
			}*/

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
	if format == "ftyp" || format == "isom" || format == "moov" || format == "mdat" || format == "moof" || format == "udta" {
		return true
	}

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

func handleIncoming(data []byte) {
	if !headerCaptured {
		idx := bytes.Index(data, []byte{0x1F, 0x43, 0xB6, 0x75})
		if idx > 0 {
			ebmlHeader = append([]byte{}, data[:idx]...)
			headerCaptured = true
		}
	} else {
		fixedData := ebmlHeader
		fixedData = append(fixedData, data...)
		data = fixedData
	}
	//fmt.Println(data)
	mutex.Lock()
	stdin.Write(data)
	mutex.Unlock()
	fmt.Println("WRITTEN TO STDIN")
	/*
		for len(data) > 0 {
			if !insideCluster {
				// Look for Cluster ID (1F 43 B6 75)
				idx := bytes.Index(data, []byte{0x1F, 0x43, 0xB6, 0x75})
				if idx == -1 {
					return // no cluster start in this chunk
				}
				insideCluster = true
				clusterBuf.Reset()
				clusterBuf.Write(data[idx:]) // start buffering
				data = data[idx+4:]          // move past ID

				var err error
				clusterSize, err = parseVint(data)
				if err != nil {
					fmt.Println("SOMETHING WENT WRONG")
					return
				}
			} else {
				// Continue buffering
				clusterBuf.Write(data)
				if clusterBuf.Len() >= clusterSize {
					// We have a full cluster → send to ffmpeg
					fmt.Println("cluster data: ", clusterBuf.Bytes())
					stdin.Write(clusterBuf.Bytes())
					insideCluster = false
					clusterBuf.Reset()
				}
				break
			}
		}
	*/
}

/*
func parseVint(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("empty data")
	}

	first := data[0]

	// Find length by checking first 1-bit
	var length int
	mask := byte(0x80) // 1000 0000
	for length = 1; length <= 8; length++ {
		if first&mask != 0 {
			break
		}
		mask >>= 1
	}
	if length > len(data) {
		return 0, fmt.Errorf("not enough bytes for VINT")
	}

	return length, nil
}
*/
