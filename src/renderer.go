package renderer

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"text/template"
)

func renderFullTemplate(w http.ResponseWriter, data map[string]any, paths ...string) {
	files := []string{"./frontend/Layout.html"}
	files = append(files, paths...)

	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		errMsg := fmt.Sprintf("Crash while rendering full templates!\nerr.Error(): %v\n", err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		fmt.Println(errMsg)
		return
	}
	
    err = tmpl.Execute(w, data)
    if err != nil {
		errMsg := fmt.Sprintf("Crash while executing full template!\nerr.Error(): %v\n", err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		fmt.Println(errMsg)
		return
	}
}

func toHTML(fileName string) string {
	return fmt.Sprintf("./frontend/%s.html", fileName)
}

func checkCookie(cookieName string, r *http.Request) (*http.Cookie, error) {
    cookie, err := r.Cookie(cookieName)
    if err != nil {
        fmt.Println("Cookie not found!")
        return nil, err
    }  
    if err = cookie.Valid(); err != nil {
        fmt.Println("Credentials invalid, please log in!")
        return nil, fmt.Errorf("Cookie was invalid!\nerr.Error(): %v\n", err.Error())
    }
    
    return cookie, nil
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
    _, err := checkCookie("user-cookie", r)
    if err != nil {
        fmt.Println(err.Error())
        renderFullTemplate(w, nil, toHTML("Login"))
        return
    }  

    http.Redirect(w, r, "/content-view", http.StatusFound)
}

func HandleViewContent(w http.ResponseWriter, r *http.Request) {
    cookie, err := checkCookie("user-cookie", r)
    if err != nil {
        fmt.Println("Credentials invalid, please log in!")        
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    userJson, err := base64.RawStdEncoding.DecodeString(cookie.Value)
    if err != nil {
        fmt.Println(err.Error())
        writeStatusMessage(w, http.StatusInternalServerError, err.Error())
        return
    }

    user := Users{}
    err = user.FromJson(userJson)
    if err != nil {
        fmt.Println(err.Error())
        writeStatusMessage(w, http.StatusInternalServerError, err.Error())
        return
    }
    con, err := getConnetion()
    if err != nil {
        fmt.Println(err.Error())
        writeStatusMessage(w, http.StatusInternalServerError, err.Error())
        return
    }
    defer con.Close()
    
    movieList, err := FetchMovieList(con)
    if err != nil {
        fmt.Println(err.Error())
        writeStatusMessage(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    content := map[string]any {
        "ContentList": movieList,
    }

    renderFullTemplate(w, content, toHTML("ContentView"))
}

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	renderFullTemplate(w, nil, toHTML("Register"))
}

func HandleRegisterContent(w http.ResponseWriter, r *http.Request) {
    cookie, err := checkCookie("user-cookie", r)
    if err != nil {
        fmt.Println("Credentials invalid, please log in!")        
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    userJson, err := base64.RawStdEncoding.DecodeString(cookie.Value)
    if err != nil {
        fmt.Println(err.Error())
        writeStatusMessage(w, http.StatusInternalServerError, err.Error())
        return
    }

    user := Users{}
    err = user.FromJson(userJson)
    if err != nil {
        fmt.Println(err.Error())
        writeStatusMessage(w, http.StatusInternalServerError, err.Error())
        return
    }
    con, err := getConnetion()
    if err != nil {
        fmt.Println(err.Error())
        writeStatusMessage(w, http.StatusInternalServerError, err.Error())
        return
    }
    defer con.Close()

    posters, err := GetSeriesPosters(con)
    if err != nil {
        fmt.Println(err.Error())
        writeStatusMessage(w, http.StatusInternalServerError, err.Error())
        return
    }

    content := map[string]any{
        "SeriesPosters": posters,
    }
    
    renderFullTemplate(w, content, toHTML("ContentRegister"))
}
