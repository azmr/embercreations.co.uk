package main

import (
	"fmt"
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	siteName string = "ember creations"
)

var templates = template.Must(template.ParseFiles("tmpl/default.html"))

type Page struct {
	Title      string        // ember creations - Page Title
	//~ Navigation template.HTML // 
	Body       template.HTML // Converted to HTML from markdown
	Css        template.CSS  // Read from tmpl/style.css
	LogoColour string        // Set here to change svg
}

func loadPage(title string) (*Page, error) {
	// TODO: add goroutines?
	logoColour := "#222"
	//~ instead of `title = strings.ToLower(title)`, TODO: redirect?
	filename := title + ".md"
	contentMD_B, err := ioutil.ReadFile("data/" + filename)
	if err != nil {
		return nil, err
	}
	
	contentHTML := template.HTML(blackfriday.MarkdownCommon(contentMD_B)) //turns MD bytes into HTML
	body := contentHTML
	css_B, err := ioutil.ReadFile("tmpl/style.css")
	if err != nil {
		log.Println(err)
	}
	css := template.CSS(css_B)

	title = siteName + " - " + strings.Title(title)
	return &Page{Title: title, Body: body, Css: css, LogoColour: logoColour}, nil
}

func renderTemplate(resp http.ResponseWriter, tmpl string, pg *Page) {
	err := templates.ExecuteTemplate(resp, tmpl+".html", pg)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
	}
}

//~ type resource struct {
//~ static bool
//~ path string
//~ }

func handler(resp http.ResponseWriter, req *http.Request) {
	// TODO: add proper redirect for odd url cases
	path := req.URL.Path[1:]
	static := false
	plainText := false
	lenPath := len(path)
	if lenPath >= 5 {
		// if path[lenPath-4:] == ".txt" redirect to ".text"
		if path[lenPath-5:] == ".text" {// TODO: pointer?
			plainText = true
			path = "data/"+path[:lenPath-5]+".md"
			fmt.Println(path)
		}
	}
	// array or slice? // others? // TODO: take filename from files in 'static' folder
	staticResources := [...]string{"favicon.ico", "robots.txt", "sitemap.xml"}
	for _, staticResource := range staticResources { // checks is request is for a static resource
		if path == staticResource {
			static = true
			break
		}
	}
	if static || plainText{ // slight inefficiency?
		http.ServeFile(resp, req, path)
	} else {
		if len(path) < 1 {
			path = "home"
		}
		pg, err := loadPage(path)
		if err != nil {
			http.Error(resp, "404 Error - File not found.", http.StatusNotFound)
			return
		}
		renderTemplate(resp, "default", pg)
	}
}

func configLogger(filename string, prefix string, flags int) {
	file, _ := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	//defer file.Close() Doesn't work with this uncommented... why?
	log.SetOutput(file)
	log.SetPrefix(prefix)
	log.SetFlags(flags)
}

func main() {
	configLogger("log.txt", "", log.Ldate|log.Ltime|log.Lshortfile)
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe("embercreations.co.uk:8080", nil))
}
