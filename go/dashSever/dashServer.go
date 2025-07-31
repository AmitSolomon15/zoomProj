package main

import (
	"bytes"
	"encoding/binary"
	"fmt"

	//"io/ioutil"
	"net/http"
	"sync"
	"time"
	//"github.com/gordonklaus/portaudio"
)

const sampleRate = 44100
const seconds = 1

var (
	buffer      = make([]float32, sampleRate*seconds)
	bufferMutex sync.RWMutex
)

func main() {
	//portaudio.Initialize()
	//defer portaudio.Terminate()

	go startServer()

	// בגלל שאין כלום בסלקט התכנית תרוץ לנצח
	select {}
}

func startServer() {
	http.HandleFunc("/audio", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Connection", "Keep-Alive")
		w.Header().Set("Transfer-Encoding", "chunked")

		// ממשיך לשלוח ללקוח אודיו כל עוד הוא לא התנתק
		for {
			bufferMutex.RLock()
			var buf bytes.Buffer
			//קורא מהבאפר לתוך משתנה באפ
			err := binary.Write(&buf, binary.BigEndian, buffer)
			bufferMutex.RUnlock()
			if err != nil {
				fmt.Println("Error writing binary:", err)
				break
			}

			_, err = w.Write(buf.Bytes())
			if err != nil {
				break // הלקוח התנתק
			}

			flusher.Flush()
			time.Sleep(time.Millisecond * 100) // המידע(שמע במקרה הזה) נשלח כל 100 מילישניות
		}
	})
	fmt.Println("HTTP server listening on :8080")
	http.ListenAndServe(":8080", nil)
}
