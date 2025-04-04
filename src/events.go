package renderer

import (
	"context"
	"database/sql"
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

const defaultPfpName string = "default_pfp.png"
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

func uploadImage(file multipart.File, filename string) (*uploader.UploadResult) {
	resp, err := images.Uploader.Upload(context.Background(), file, uploader.UploadParam{
		FileName: filename,
	})
	if err != nil {
		return nil
	}
    return &resp.Data
}

func deleteImage(fileId string) error {
    _, err := images.Media.DeleteFile(context.Background(), fileId)
    return err
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

	user := users{
		Username: r.PostFormValue("username"),
		Email:    r.PostFormValue("email"),
		Password: hash(r.PostFormValue("password")),
    }

    err = user.Insert(con)
    if err != nil {
        fmt.Println(err.Error())
        writeStatusMessage(w, http.StatusBadRequest, err.Error())
        w.WriteHeader(http.StatusBadRequest)
        return
    }

	r.ParseMultipartForm(10 << 20)
	file, header, err := r.FormFile("pfp")
	if err != nil {
        fmt.Printf("\nUser: %s, registered without pfp!\nerr.Error(): %s", user.Username, err.Error())
        writeStatusMessage(w, http.StatusOK, "User registered without pfp!")
	    return
    }
	defer file.Close()
    imgUrl := "https://ik.imagekit.io/FilmPost/default_pfp.png?updatedAt=1743482764769"

    imgResult := uploadImage(file, header.Filename)
    if imgResult != nil {
        imgUrl = imgResult.Url
    }
    
    user.PfpUrl = sql.NullString{ String: imgUrl, Valid: len(imgUrl) > 0 } 
    err = user.Update(con)
    if err != nil {
        errMsg := fmt.Sprintf("Crash while setting the user pfp!\n%v\n", err.Error())
    
        fmt.Println(errMsg)
        writeStatusMessage(w, http.StatusBadRequest, errMsg)
        return
    }

    fmt.Printf("\nUser: %s, registered succesfully!\n", user.Username)
	writeStatusMessage(w, http.StatusOK, "User succesfully registered!")
}
