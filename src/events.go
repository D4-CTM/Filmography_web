package renderer

import (
	"context"
	"database/sql"
	"encoding/base64"
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
)

const (
	DEFAULT_SERIES_POSTER_NAME string = "Add new series"
	DEFAULT_MOVIE_POSTER       string = "https://ik.imagekit.io/FilmPost/movies.svg?updatedAt=1743831041071"
	DEFAULT_SERIES_POSTER      string = "https://ik.imagekit.io/FilmPost/series.svg?updatedAt=1743831042090"
	DEFAULT_USER_PFP           string = "https://ik.imagekit.io/FilmPost/default_pfp.png?updatedAt=1743482764769"
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

func uploadImage(file multipart.File, filename string) *uploader.UploadResult {
	resp, err := images.Uploader.Upload(context.Background(), file, uploader.UploadParam{
		FileName: filename,
	})
	if err != nil {
		fmt.Println(err.Error())
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
		PfpUrl:   sql.NullString{String: DEFAULT_USER_PFP, Valid: true},
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
	imgUrl := DEFAULT_USER_PFP

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
	w.Header().Add("HX-Location", "/login")
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

	userJson, err := user.ToJson()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		fmt.Println(err.Error())
		return
	}

	cookie := http.Cookie{
		Name:     "user-cookie",
		Value:    base64.RawStdEncoding.EncodeToString(userJson),
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	fmt.Println("User logged in!")
	w.Header().Set("HX-Location", "/content-view")
	writeStatusMessage(w, http.StatusOK, "User logged in succesfully!")
}

func insertMovie(w http.ResponseWriter, r *http.Request, movie Movies) error {
	con, err := getConnetion()
	if err != nil {
		return err
	}
	defer con.Close()

	err = movie.Insert(con)
	if err != nil {
		return err
	}

	r.ParseMultipartForm(10 << 20)
	file, header, err := r.FormFile("poster")
	if err != nil {
		fmt.Printf("\nmovie: %s, registered without poster!\nerr.Error(): %s", movie.Name, err.Error())
		writeStatusMessage(w, http.StatusCreated, "Movie registered without poster!")
		return nil
	}
	defer file.Close()

	imgUrl := uploadImage(file, header.Filename)
	if imgUrl == nil {
		return fmt.Errorf("Crash on poster upload!")
	}

	fmt.Printf("\n\timgUrl: %s\n", imgUrl.Url)
	movie.PosterUrl = sql.NullString{String: imgUrl.Url, Valid: len(imgUrl.Url) > 0}
	fmt.Printf("\n\tMovie content: %v\n\tPoster: %v\n", movie, movie.PosterUrl)
	err = movie.Update(con)
	if err != nil {
		errMsg := fmt.Sprintf("\nCrash while inserting the movie poster!\nerr.Error(): %v\n", err.Error())
		fmt.Println(errMsg)
		return fmt.Errorf(errMsg)
	}

	writeStatusMessage(w, http.StatusCreated, "Movie registered succesfully!")
	return nil
}

func insertEpisode(w http.ResponseWriter, r *http.Request, episode SeriesEpisode) error {
	con, err := getConnetion()
	if err != nil {
		return err
	}
	defer con.Close()

	err = episode.Insert(con)
	if err != nil {
		return err
	}

	seriesPosterName := r.PostFormValue("series-poster")

	// We simply check if the series is specified
	if len(seriesPosterName) == 0 {
		writeStatusMessage(w, http.StatusCreated, "Episode registered without a poster!")
		return nil
	}

    seriesName := r.PostFormValue("series-name")
    if len(seriesName) == 0 {
		writeStatusMessage(w, http.StatusCreated, "Episode registered without a poster!")
		return nil 
    }
    episode.Poster.SeriesName = sql.NullString{String: seriesName, Valid: true}

	r.ParseMultipartForm(10 << 20)
	file, header, err := r.FormFile("poster")
	if err != nil {
        if seriesPosterName != DEFAULT_SERIES_POSTER_NAME {
            err = episode.Update(con)
            if err != nil {
                errMsg := fmt.Sprintf("\nCrash while specifying the series!\nerr.Error(): %v\n", err.Error())
                fmt.Println(errMsg)
                return fmt.Errorf(errMsg)
            }

            writeStatusMessage(w, http.StatusCreated, "Episode registered succesfully!")
            return nil
        }

        fmt.Printf("\nEpisode: %s, registered without poster!\nerr.Error(): %s", episode.Serie.Name, err.Error())
		writeStatusMessage(w, http.StatusCreated, "Episode registered without poster!")
		return nil
	}
	defer file.Close()

	imgUrl := uploadImage(file, header.Filename)
	if imgUrl == nil {
		return fmt.Errorf("Crash on poster upload!")
	}

	fmt.Printf("\n\timgUrl: %s\n", imgUrl.Url)
	episode.Poster.PosterUrl = sql.NullString{String: imgUrl.Url, Valid: len(imgUrl.Url) > 0}
    episode.Poster.SeriesName = sql.NullString{String: seriesName, Valid: true}
    fmt.Printf("\n\tEpisode content: %v\n\tPoster: %v\n", episode, episode.Poster.PosterUrl)
	err = episode.Update(con)
	if err != nil {
		errMsg := fmt.Sprintf("\nCrash while inserting the series poster!\nerr.Error(): %v\n", err.Error())
		fmt.Println(errMsg)
		return fmt.Errorf(errMsg)
	}

	writeStatusMessage(w, http.StatusCreated, "Episode & serie registered succesfully!")
	return nil
}

func EventRegisterContent(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Starting content register process!")

	contentName := r.PostFormValue("name")
	description := r.PostFormValue("description")
	rating, err := strconv.Atoi(r.PostFormValue("rating"))
	if err != nil {
		errMsg := fmt.Sprintf("Crash while converting the ratins!\nerr.Error(): %v\n", err.Error())
		fmt.Println(errMsg)
		writeStatusMessage(w, http.StatusBadRequest, errMsg)
		return
	}
	cookie, err := checkCookie("user-cookie", r)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user Users
	userJson, err := base64.RawStdEncoding.DecodeString(cookie.Value)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = user.FromJson(userJson)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	insertionElement := r.PostFormValue("content")

	switch insertionElement {
	case "M":
		movie := Movies{
			Name:        contentName,
			Description: sql.NullString{String: description, Valid: len(description) > 0},
			Stars:       int16(rating),
			AddedBy:     user.Id,
			PosterUrl:   sql.NullString{String: DEFAULT_MOVIE_POSTER, Valid: true},
		}

		err = insertMovie(w, r, movie)
		if err != nil {
			fmt.Println(err.Error())
			writeStatusMessage(w, http.StatusBadRequest, err.Error())
		}

	case "S":
		episode := SeriesEpisode{
			Serie: Episode{
				Name:        contentName,
				Description: sql.NullString{String: description, Valid: len(description) > 0},
				Stars:       int16(rating),
				AddedBy:     user.Id,
			},
			Poster: SeriesPoster{
				SeriesName: sql.NullString{String: DEFAULT_SERIES_POSTER_NAME, Valid: true},
				PosterUrl:  sql.NullString{String: DEFAULT_SERIES_POSTER, Valid: true},
			},
		}

		err = insertEpisode(w, r, episode)
		if err != nil {
			fmt.Println(err.Error())
			writeStatusMessage(w, http.StatusBadRequest, err.Error())
		}
	default:
		fmt.Println("Didn't select any!")
		writeStatusMessage(w, http.StatusBadRequest, "Please select what type of content are you rating!")
	}
	fmt.Println("Finish insertion process content!")
}
