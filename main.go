package main

import (
	"log"
	"net/http"

	renderer "main.go/src"
)

func main() {
	err := renderer.InitDotEnv()
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = renderer.InitImageKet()
	if err != nil {
		log.Fatalln(err.Error())
	}

	log.Println("Starting server...")
	static := http.FileServer(http.Dir("backend"))

	mux := http.NewServeMux()
	mux.Handle("/backend/", http.StripPrefix("/backend/", static))
	//functions of the backend
    mux.HandleFunc("/login-user/", renderer.EventLogin)
    mux.HandleFunc("/register-user/", renderer.EventRegisterUser)
    mux.HandleFunc("/register-content/", renderer.EventRegisterContent)
	// actual sites on the page
    mux.HandleFunc("/login", renderer.HandleLogin)
    mux.HandleFunc("/register", renderer.HandleRegister)
    mux.HandleFunc("/content-view", renderer.HandleViewContent)
	mux.HandleFunc("/content-register", renderer.HandleRegisterContent)

	log.Println("Server started!")
	log.Fatal(http.ListenAndServe(":5412", mux))
}
