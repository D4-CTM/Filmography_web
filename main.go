package main

import (
	"fmt"
	"log"
	"net/http"

	renderer "main.go/src"
)

func main() {
	err := renderer.InitDotEnv()
	if err != nil {
		log.Fatalf("Dotnev couldn't be found\nerr.Error(): %v\n",err.Error())
	}

	err = renderer.InitImageKit()
	if err != nil {
		log.Fatalf("Image Kit couldn't be loaded!\nerr.Error(): %v\n", err.Error())
	}

	log.Println("Starting server...")
	static := http.FileServer(http.Dir("backend"))

	mux := http.NewServeMux()
	mux.Handle("/backend/", http.StripPrefix("/backend/", static))
	//functions of the backend
	mux.HandleFunc("/logout/", renderer.EventLogOut)
	mux.HandleFunc("/login-user/", renderer.EventLogin)
	mux.HandleFunc("/register-user/", renderer.EventRegisterUser)
	mux.HandleFunc("/register-content/", renderer.EventRegisterContent)
	mux.HandleFunc("/episode/delete/{content}/", renderer.EventDeleteEpisode)
	mux.HandleFunc("/movie/delete/{content}/", renderer.EventDeleteMovie)
	// actual sites on the page
	mux.HandleFunc("/login", renderer.HandleLogin)
	mux.HandleFunc("/register", renderer.HandleRegister)
	mux.HandleFunc("/content-view", renderer.HandleViewContent)
	mux.HandleFunc("/content-register", renderer.HandleRegisterContent)

	log.Println("Server started!")
	fmt.Printf("Connect to http://localhost:5412/\n")
	log.Fatal(http.ListenAndServe(":5412", mux))
}
