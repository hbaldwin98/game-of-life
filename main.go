package main

import (
	"html/template"
	"log"
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
	router.HandleFunc("/index", indexHandler)
	router.HandleFunc("/board", boardHandler)
	router.HandleFunc("/genetic", geneticHandler)
	router.HandleFunc("/place", placeHandler)

	// Middleware
	logger := log.New(os.Stdout, "", log.LstdFlags)
	loggingMiddleware := LoggingMiddleware(logger)

	// Router, wrapped with middleware
	configuredRouter := loggingMiddleware(router)

	log.Printf("Server started in %s", time.Since(timer))
	if err := http.ListenAndServe(":8080", configuredRouter); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
		os.Exit(1)
	}
}

func LoggingMiddleware(logger *log.Logger) func(http.Handler) http.Handler {
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

	t, _ := template.ParseFiles("board.html")

	t.ExecuteTemplate(w, "board", data)
}

type GeneticData struct {
	Board      [][]bool
	Generation int
}

var geneticBoard GeneticData
var boardSize int = 40

func geneticHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/genetic" {
		http.NotFound(w, r)
		return
	}

	if geneticBoard.Board == nil {
		geneticBoard.Board = make([][]bool, boardSize)
		for i := range geneticBoard.Board {
			geneticBoard.Board[i] = make([]bool, boardSize)
		}
	}

	if r.URL.Query().Get("update") == "true" {
		geneticBoard.NextGeneration()
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

func (g *GeneticData) NextGeneration() {
	newBoard := make([][]bool, boardSize)
	for i := range newBoard {
		newBoard[i] = make([]bool, boardSize)
	}

	for i, gene := range g.Board {
		for j, alive := range gene {

			neighbors := g.GetNumberNeighbors(i, j)

			if alive {
				if neighbors < 2 || neighbors > 3 {
					newBoard[i][j] = false
				} else {
					newBoard[i][j] = true
				}
			} else {
				if neighbors == 3 {
					newBoard[i][j] = true
				} else {
					newBoard[i][j] = false
				}
			}
		}
	}

	for i := range g.Board {
		copy(g.Board[i], newBoard[i])
	}
}

func (g *GeneticData) GetNumberNeighbors(x, y int) int {
	count := 0

	neighbors := [][2]int{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}

	for _, offset := range neighbors {
		nX, nY := x+offset[0], y+offset[1]
		if nX >= 0 && nX < boardSize && nY >= 0 && nY < boardSize && g.GetGene(nX, nY) {
			count++
		}
	}

	return count
}

func (g *GeneticData) GetGene(x, y int) bool {
	if x < 0 || x >= boardSize || y < 0 || y >= boardSize {
		return false
	}

	return g.Board[x][y]
}

func parseTemplate(filename string) *template.Template {
	return template.Must(template.ParseFiles(filename))
}
