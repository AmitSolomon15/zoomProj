package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	//"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	client *mongo.Client
	ctx    context.Context
)

func main() {
	setClient()
	defer client.Disconnect(ctx)

	http.HandleFunc("/submit-data-Sign-Up", submitHandler)
	http.HandleFunc("/submit-data-Sign-In", signInHandler)
	http.HandleFunc("/get-users", getUsers)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local testing
	}
	http.ListenAndServe(":"+port, nil)
}

func assignPort() {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}
	defer listener.Close()
	// Get the assigned address and port
	addr := listener.Addr().(*net.TCPAddr)
	fmt.Printf("Server listening on IP: %s, Port: %d\n", addr.IP.String(), addr.Port)
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
	fName := r.FormValue("fName")
	lName := r.FormValue("lName")
	uName := r.FormValue("uName")
	pass := r.FormValue("pass")

	//creates database
	usersDatabase := client.Database("users")
	//create collection in the database
	userCollection := usersDatabase.Collection("users")

	//insert data of 1 object

	userResult, err := userCollection.InsertOne(ctx, bson.D{
		{Key: "first name", Value: fName},
		{Key: "last name", Value: lName},
		{Key: "username", Value: uName},
		{Key: "password", Value: pass},
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
	uName := r.FormValue("uName")
	pass := r.FormValue("pass")

	//creates database
	usersDatabase := client.Database("users")
	//create collection in the database
	userCollection := usersDatabase.Collection("users")

	cursor, err := userCollection.Find(ctx, bson.M{"username": uName, "password": pass})

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

func getUsers(w http.ResponseWriter, r *http.Request) {

	setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	//creates database
	usersDatabase := client.Database("users")
	//create collection in the database
	userCollection := usersDatabase.Collection("users")

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

//func sendName(w http.ResponseWriter){
//
//}
