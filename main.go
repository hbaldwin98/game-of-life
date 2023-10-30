package main

import (
	"html/template"
	"log"
	"main/conway"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	log.Println("Starting server...")
	timer := time.Now()

	router := http.NewServeMux()
	router.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	// Routes
	router.Handle("/", http.RedirectHandler("/index", http.StatusFound))
	router.HandleFunc("/index", indexHandler)
	router.HandleFunc("/board", boardHandler)
	router.HandleFunc("/genetic", geneticHandler)
	router.HandleFunc("/place", placeHandler)

	// Middleware
	logger := log.New(os.Stdout, "", log.LstdFlags)
	loggingMiddleware := loggingMiddleware(logger)

	// Router, wrapped with middleware
	configuredRouter := loggingMiddleware(router)

	log.Printf("Server started on localhost:8080 in %s", time.Since(timer))
	if err := http.ListenAndServe(":8080", configuredRouter); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
		os.Exit(1)
	}
}

func loggingMiddleware(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logger.Println(err)
				}
			}()

			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
		}

		return http.HandlerFunc(fn)
	}
}

type PageData struct {
	Title string
	Body  string
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/index" {
		http.NotFound(w, r)
		return
	}

	data := PageData{
		Title: "Game of Life",
		Body:  "Welcome to the Game of Life!",
	}
	t, _ := template.ParseFiles("public/index.html")

	t.ExecuteTemplate(w, "index", data)
}

type BoardState struct {
	Time  int
	Pause bool
	Next  bool
}

func boardHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/board" {
		http.NotFound(w, r)
		return
	}

	query := r.URL.Query()
	time := query.Get("tick")
	pause := query.Get("pause")
	next := query.Get("next")

	// ticks per second, convert to milliseconds
	// 1 = 1000
	// 2 = 500
	data := BoardState{
		Time:  1000,
		Pause: pause == "true",
		Next:  next == "true",
	}

	if time != "" {
		timeInSeconds, err := strconv.Atoi(time)

		if err != nil {
			timeInSeconds = 1
		}

		data.Time = 1000 / timeInSeconds
	}

	t := parseTemplate("board.html")

	t.ExecuteTemplate(w, "board", data)
}

var BOARD_SIZE int = 40
var geneticBoard conway.GeneticData

func geneticHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/genetic" {
		http.NotFound(w, r)
		return
	}

	if geneticBoard.Board == nil {
		geneticBoard.Board = make([][]bool, BOARD_SIZE)
		for i := range geneticBoard.Board {
			geneticBoard.Board[i] = make([]bool, BOARD_SIZE)
		}
	}

	if r.URL.Query().Get("update") == "true" {
		geneticBoard.NextGeneration(BOARD_SIZE)
		geneticBoard.Generation++
	}

	t := parseTemplate("genetic.html")

	if err := t.ExecuteTemplate(w, "genetic", &geneticBoard); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func placeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/place" {
		http.NotFound(w, r)
		return
	}

	query := r.URL.Query()
	x := query.Get("x")
	y := query.Get("y")

	posX, err := strconv.Atoi(x)
	posY, err := strconv.Atoi(y)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	existingGene := geneticBoard.Board[posX][posY]

	if existingGene {
		geneticBoard.Board[posX][posY] = false
	} else {
		geneticBoard.Board[posX][posY] = true
	}

	t := parseTemplate("genetic.html")

	if err := t.ExecuteTemplate(w, "genetic", &geneticBoard); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func parseTemplate(filename string) *template.Template {
	return template.Must(template.ParseFiles(filename))
}
