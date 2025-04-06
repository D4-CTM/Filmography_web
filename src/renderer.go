package renderer

import (
	"fmt"
	"net/http"
	"text/template"
)

func renderFullTemplate(w http.ResponseWriter, content map[string]any, paths ...string) {
	files := []string{"./frontend/Layout.html"}
	files = append(files, paths...)

	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		errMsg := fmt.Sprintf("Crash while rendering full templates!\nerr.Error(): %v\n", err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		fmt.Println(errMsg)
		return
	}
	err = tmpl.Execute(w, content)
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

func HandleViewContent(w http.ResponseWriter, r *http.Request) {
    renderFullTemplate(w, nil, toHTML("ContentView"))
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
    renderFullTemplate(w, nil, toHTML("Login"))
}

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	renderFullTemplate(w, nil, toHTML("Register"))
}

func HandleRegisterContent(w http.ResponseWriter, r *http.Request) {
	renderFullTemplate(w, nil, toHTML("ContentRegister"))
}
