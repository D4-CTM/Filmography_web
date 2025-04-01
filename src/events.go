package renderer

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/imagekit-developer/imagekit-go"
	"github.com/imagekit-developer/imagekit-go/api/uploader"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var images *imagekit.ImageKit

func simpleHash(str string) int {
	result := 1
	runes := []rune(str)
	for i := range len(str) {
		result = result + (int(runes[i]) * len(str))
	}
	return result
}

func hash(str string) int {
	key := simpleHash(os.Getenv("SECRET_PASSWORD"))
	runes := []rune(str)
	result := 1
	for i := range len(str) {
		result = result + (int(runes[i]))*key
	}
	return result
}

func InitImageKet() error {
	ik, err := imagekit.New()
	if err != nil {
		return fmt.Errorf("There was an error initializing the image kit!\nerr.Error(): %v\n", err.Error())
	}
	images = ik
	return nil
}

func InitDotEnv() error {
	return godotenv.Load()
}

func getConnetion() (*sqlx.DB, error) {
	con, err := sqlx.Open("postgres", os.Getenv("CONNECTION_STRING"))
	if err != nil {
		return nil, fmt.Errorf("Crash while stablishing conection!\nerr.Error(): %v\n", err.Error())
	}
	return con, nil
}

func uploadImage(file multipart.File, filename string) (Url string) {
	resp, err := images.Uploader.Upload(context.Background(), file, uploader.UploadParam{
		FileName: filename,
	})
	if err != nil {
        return "https://ik.imagekit.io/FilmPost/default_pfp.png?updatedAt=1743482764769" 
	}
    return resp.Data.Url
}

func writeStatusMessage(w http.ResponseWriter, status int, message string) {
	w.Header().Set("HX-Status", fmt.Sprint(status))
	w.Header().Set("HX-Message", message)
	w.WriteHeader(status)
}

func EventRegisterUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Beginning user registration...")

	con, err := getConnetion()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		fmt.Println(err.Error())
		return
	}
	defer con.Close()

	r.ParseMultipartForm(10 << 20)
	file, header, err := r.FormFile("pfp")
	if err != nil {
		fmt.Printf("Crash while fetching the pfp!\nerr.Error(): %v\n", err.Error())
        writeStatusMessage(w, http.StatusBadRequest, fmt.Sprintf("There was an error loading the image file! err.Error(): %v", err.Error()))
        w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer file.Close()

	user := users{
		Username: r.PostFormValue("username"),
		Email:    r.PostFormValue("email"),
		Password: hash(r.PostFormValue("password")),
		PfpUrl:   uploadImage(file, header.Filename),
	}

	fmt.Printf("\nUser: %s, registered succesfully!\n", user.Username)
	writeStatusMessage(w, http.StatusOK, "User succesfully registered!")
}
