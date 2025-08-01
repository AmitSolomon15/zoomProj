package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	//"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	http.HandleFunc("/submit-data-Sign-Up", submitHandler)
	http.HandleFunc("/submit-data-Sign-In", signInHandler)
	http.HandleFunc("/get-users", getUsers)
	http.ListenAndServe("https://zoomproj-back.onrender.com", nil)
}

// sign up
func submitHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

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

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://amitsol462:AmitS210706@cluster0.jbild9v.mongodb.net/"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	//connect user to data base
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(ctx)

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
		log.Fatal(err)
	}

	fmt.Println(userResult.InsertedID)

	cursor, err := userCollection.Find(ctx, bson.M{})

	if err != nil {
		log.Fatal(err)
	}

	var users []bson.M
	if err = cursor.All(ctx, &users); err != nil {
		log.Fatal(err)
	}

	for _, user := range users {
		fmt.Println(user["username"])
	}

}

// sign in
func signInHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		return // preflight
	}

	// Parse the form data
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error parsing form", http.StatusInternalServerError)
		return
	}

	// Get values from the form
	uName := r.FormValue("uName")
	pass := r.FormValue("pass")

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://amitsol462:AmitS210706@cluster0.jbild9v.mongodb.net/"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	//connect user to data base
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(ctx)

	//creates database
	usersDatabase := client.Database("users")
	//create collection in the database
	userCollection := usersDatabase.Collection("users")

	cursor, err := userCollection.Find(ctx, bson.M{"username": uName, "password": pass})

	if err != nil {
		log.Fatal(err)
	}

	var users []bson.M
	if err = cursor.All(ctx, &users); err != nil {
		log.Fatal(err)
	}
	for _, user := range users {
		fmt.Println(user["username"])
	}

	if len(users) > 0 {
		json.NewEncoder(w).Encode(users[0]) // Send one user as JSON
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"user not found"}`))
	}
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		return // preflight
	}

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://amitsol462:AmitS210706@cluster0.jbild9v.mongodb.net/"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	//connect user to data base
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(ctx)

	//creates database
	usersDatabase := client.Database("users")
	//create collection in the database
	userCollection := usersDatabase.Collection("users")

	cursor, err := userCollection.Find(ctx, bson.M{})

	if err != nil {
		log.Fatal(err)
	}

	var users []bson.M
	if err = cursor.All(ctx, &users); err != nil {
		log.Fatal(err)
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
