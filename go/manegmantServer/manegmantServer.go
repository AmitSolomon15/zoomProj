package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	//"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	client   *mongo.Client
	ctx      context.Context
	listener *net.UDPConn
	user     mUser
)

type mUser struct {
	fName string
	lName string
	uName string
	pass  string
	port  string
	ip    string
}

func main() {
	setClient()
	defer client.Disconnect(ctx)

	http.HandleFunc("/submit-data-Sign-Up", submitHandler)
	http.HandleFunc("/submit-data-Sign-In", signInHandler)
	http.HandleFunc("/disconnect", disconnectHandler)
	http.HandleFunc("/get-users", getUsers)
	http.HandleFunc("/connect-user-udp", startChatHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local testing
	}
	http.ListenAndServe(":"+port, nil)
}

func assignPort() {
	var err error

	listener, err = net.ListenUDP("udp", &net.UDPAddr{Port: 0})
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}

	//creates database
	usersDatabase := client.Database("users")
	//create collection in the database
	userOnLineCollection := usersDatabase.Collection("usersOnLine")
	defer listener.Close()
	// Get the assigned address and port
	addr := listener.LocalAddr().(*net.UDPAddr)
	user.port = fmt.Sprint(addr.Port)

	ip, err := getOutboundIP()
	if err != nil {
		log.Printf("Could not get outbound IP: %v", err)
		return
	} else {
		user.ip = ip.String()
	}

	userOnLineRes, err := userOnLineCollection.InsertOne(ctx, bson.D{
		{Key: "username", Value: user.uName},
		{Key: "ip", Value: user.ip},
		{Key: "port", Value: fmt.Sprint(addr.Port)},
		{Key: "time", Value: time.Now()},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(userOnLineRes)
	fmt.Printf("Server listening on IP: %s, Port: %d\n", addr.IP.String(), addr.Port)

	go func() {
		buf := make([]byte, 1024)
		for {
			n, addr, err := listener.ReadFromUDP(buf)
			if err != nil {
				log.Printf("Read error: %v", err)
				continue
			}
			fmt.Printf("Received from %s: %s\n", addr.String(), string(buf[:n]))
		}
	}()
}

func getOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}

// set headers
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "https://zoomproj-front.onrender.com")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

}

func setClient() {
	var err error
	client, err = mongo.NewClient(options.Client().ApplyURI("mongodb+srv://amitsol462:AmitS210706@cluster0.jbild9v.mongodb.net/"))
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// sign up
func submitHandler(w http.ResponseWriter, r *http.Request) {

	setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		return // preflight
	}

	// Parse the form data
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error parsing form", http.StatusInternalServerError)
		return
	}

	// Get values from the form
	user.fName = r.FormValue("fName")
	user.lName = r.FormValue("lName")
	user.uName = r.FormValue("uName")
	user.pass = r.FormValue("pass")

	//creates database
	usersDatabase := client.Database("users")
	//create collection in the database
	userCollection := usersDatabase.Collection("users")

	//insert data of 1 object

	userResult, err := userCollection.InsertOne(ctx, bson.D{
		{Key: "first name", Value: user.fName},
		{Key: "last name", Value: user.lName},
		{Key: "username", Value: user.uName},
		{Key: "password", Value: user.pass},
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(userResult.InsertedID)

	signInHandler(w, r)

}

// sign in
func signInHandler(w http.ResponseWriter, r *http.Request) {

	setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse the form data
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error parsing form", http.StatusInternalServerError)
		return
	}

	// Get values from the form
	user.uName = r.FormValue("uName")
	user.pass = r.FormValue("pass")

	//creates database
	usersDatabase := client.Database("users")
	//create collection in the database
	userCollection := usersDatabase.Collection("users")

	cursor, err := userCollection.Find(ctx, bson.M{"username": user.uName, "password": user.pass})

	if err != nil {
		fmt.Println(err)
		return
	}

	var users []bson.M
	if err = cursor.All(ctx, &users); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(users))
	for _, user := range users {
		fmt.Println(user["username"])
	}

	if len(users) > 0 {
		json.NewEncoder(w).Encode(users[0]) // Send one user as JSON
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"user not found"}`))
	}

	assignPort()
}

func disconnectHandler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	fmt.Println("entered disconnect")

	//fmt.Println(user.uName)
	var uName string
	fmt.Println(r.Body)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error parsing form", http.StatusInternalServerError)
		return
	}

	uName = r.FormValue("username")
	fmt.Println(uName)
	sessionCollection := client.Database("users").Collection("usersOnLine")
	_, err := sessionCollection.DeleteOne(ctx, bson.M{"username": uName})
	if err != nil {
		http.Error(w, "Failed to delete session", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Disconnected user: %s\n", uName)
	w.WriteHeader(http.StatusOK)
}

func getUsers(w http.ResponseWriter, r *http.Request) {

	setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	//creates database
	usersDatabase := client.Database("users")
	//create collection in the database
	userCollection := usersDatabase.Collection("usersOnLine")

	cursor, err := userCollection.Find(ctx, bson.M{})

	if err != nil {
		fmt.Println(err)
		return
	}

	var users []bson.M
	if err = cursor.All(ctx, &users); err != nil {
		fmt.Println(err)
		return
	}
	for _, user := range users {
		fmt.Println(user["username"])
	}

	if len(users) > 0 {
		json.NewEncoder(w).Encode(users) // Send one user as JSON
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"user not found"}`))
	}
}

func startChatHandler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	r.ParseMultipartForm(10 << 20)
	//fromUser := r.FormValue("from")
	toUser := r.FormValue("to")

	collection := client.Database("users").Collection("usersOnLine")

	var to bson.M
	if err := collection.FindOne(ctx, bson.M{"username": toUser}).Decode(&to); err != nil {
		http.Error(w, "User not found", 404)
		return
	}

	toPort := to["port"]
	toIP := to["ip"]
	message := r.Form["msg"]
	joinedMsg := strings.Join(message, "")

	raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", toIP, toPort))
	if err != nil {
		log.Println("Resolve error:", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Println("Dial error:", err)
		return
	}
	defer conn.Close()

	conn.Write([]byte(joinedMsg))
}
