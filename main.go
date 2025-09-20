package main

import (
	"embed"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// --- Embed assets ---
//go:embed templates/*
//go:embed static/*
var content embed.FS

var templates = template.Must(template.ParseFS(content, "templates/*.html"))

// --- In-memory state ---
type Room struct {
	ID   string
	Log  []string
	Lock sync.Mutex
}

var (
	rooms = map[string]*Room{}
	mu    sync.Mutex
)

// --- Handlers ---
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		roomID := strconv.FormatInt(time.Now().UnixNano(), 36)
		mu.Lock()
		rooms[roomID] = &Room{ID: roomID}
		mu.Unlock()
		http.Redirect(w, r, "/room/"+roomID, http.StatusSeeOther)
		return
	}
	templates.ExecuteTemplate(w, "index.html", nil)
}

func roomHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Path[len("/room/"):]
	mu.Lock()
	room, ok := rooms[roomID]
	mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		dice := rand.Intn(6) + 1
		entry := name + " rolled a " + strconv.Itoa(dice)

		room.Lock.Lock()
		room.Log = append(room.Log, entry)
		room.Lock.Unlock()
	}

	room.Lock.Lock()
	defer room.Lock.Unlock()
	templates.ExecuteTemplate(w, "room.html", room)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Serve static files
	fs := http.FileServer(http.FS(content))
	http.Handle("/static/", fs)

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/room/", roomHandler)

	log.Println("Listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
