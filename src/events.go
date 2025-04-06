package renderer

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/imagekit-developer/imagekit-go"
	"github.com/imagekit-developer/imagekit-go/api/uploader"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
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

func uploadImage(file multipart.File, filename string) *uploader.UploadResult {
	resp, err := images.Uploader.Upload(context.Background(), file, uploader.UploadParam{
		FileName: filename,
	})
	if err != nil {
		return nil
	}
	return &resp.Data
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

	user := Users{
		Username: r.PostFormValue("username"),
		Email:    r.PostFormValue("email"),
		Password: hash(r.PostFormValue("password")),
		PfpUrl:   sql.NullString{String: "https://ik.imagekit.io/FilmPost/default_pfp.png?updatedAt=1743482764769", Valid: true},
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

	user.PfpUrl = sql.NullString{String: imgUrl, Valid: len(imgUrl) > 0}
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

func EventLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Starting login auth!")
	con, err := getConnetion()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		fmt.Println(err.Error())
		return
	}
	defer con.Close()

	user := Users{
		Username: r.PostFormValue("username"),
        Password: hash(r.PostFormValue("password")),
	}
    err = user.Fetch(con)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		fmt.Println(err.Error())
		return
	}
   
    json, err := json.Marshal(user)    
    if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		fmt.Println(err.Error())
        return
    }

    cookie := http.Cookie{
        Name: "user-cookie",
        Value: base64.RawStdEncoding.EncodeToString(json),
        Secure: true,
        HttpOnly: true,
    }
    http.SetCookie(w, &cookie)

	fmt.Println("User logged in!")
    w.Header().Set("HX-Location", "/content-view")
	writeStatusMessage(w, http.StatusAccepted, "User logged in!")
}

func insertMovie(w http.ResponseWriter, r *http.Request, movie Movies) error {
	con, err := getConnetion()
	if err != nil {
		return err
	}
	defer con.Close()
	movie.PosterUrl = sql.NullString{String: "https://ik.imagekit.io/FilmPost/movies.svg?updatedAt=1743831041071", Valid: true}

	err = movie.Insert(con)

	r.ParseMultipartForm(10 << 20)
	file, header, err := r.FormFile("pfp")
	if err != nil {
		fmt.Printf("\nmovie: %s, registered without poster!\nerr.Error(): %s", movie.Name, err.Error())
		writeStatusMessage(w, http.StatusOK, "Movie registered without poster!")
		return nil
	}
	defer file.Close()
	imgUrl := "https://ik.imagekit.io/FilmPost/movies.svg?updatedAt=1743831041071"

	imgResult := uploadImage(file, header.Filename)
	if imgResult != nil {
		imgUrl = imgResult.Url
	}

	movie.PosterUrl = sql.NullString{String: imgUrl, Valid: len(imgUrl) > 0}
	err = movie.Update(con)
	if err != nil {
		fmt.Printf("\nCrash while inserting the movie poster!\nerr.Error(): %v\n", err.Error())
	}
	return nil
}

func EventRegisterContent(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Starting content register!")

	contentName := r.PostFormValue("name")
	description := r.PostFormValue("description")
	rating, err := strconv.Atoi(r.PostFormValue("rating"))
	if err != nil {
		errMsg := fmt.Sprintf("Crash while converting the ratins!\nerr.Error(): %v\n", err.Error())
		fmt.Println(errMsg)
		writeStatusMessage(w, http.StatusBadRequest, errMsg)
		return
	}
    cookie, err := r.Cookie("user-cookie")
    if err != nil {
        fmt.Println(err.Error())
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    var user Users
    json, err := base64.RawStdEncoding.DecodeString(cookie.Value)
    if err != nil {
        fmt.Println(err.Error())
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    

	insertionElement := r.PostFormValue("content")
	if insertionElement == "M" {

		movie := Movies{
			Name:        contentName,
			Description: sql.NullString{String: description, Valid: len(description) > 0},
			Stars:       int16(rating),
            AddedBy:    ,
		}

		insertMovie(w, r, movie)

	} else if insertionElement == "S" {

	} else {
		fmt.Println("Didn't select any!")
		writeStatusMessage(w, http.StatusBadRequest, "Please select what type of content are you rating!")
		return
	}
}
