package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// --- Embed assets ---
//
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
	rooms      = map[string]*Room{}
	mu         sync.Mutex
	hostPrefix string
)

// --- Handlers ---
func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("indexHandler: %s %s", r.Method, r.URL.String())
	if r.Method == http.MethodPost {
		log.Printf("POST method recieved'....")
		// Generate unique room ID
		roomID := strconv.FormatInt(time.Now().UnixNano(), 36)

		mu.Lock()
		rooms[roomID] = &Room{ID: roomID}
		mu.Unlock()

		http.Redirect(w, r, hostPrefix+"/room/"+roomID, http.StatusSeeOther)
		return
	}

	if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func roomHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("roomHandler: %s %s", r.Method, r.URL.String())
	roomID := r.URL.Path[len("/room/"):]
	log.Printf("room id is %s", roomID)

	mu.Lock()
	room, ok := rooms[roomID]
	mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	room.Lock.Lock()
	defer room.Lock.Unlock()

	// Check if user already has a name in cookie
	var userName string
	cookie, err := r.Cookie("username")
	if err == nil {
		userName = cookie.Value
	}

	if userName == "" {
		// No name yet
		if r.Method == http.MethodPost {
			userName = r.FormValue("name")
			if userName == "" {
				http.Error(w, "Name required", http.StatusBadRequest)
				return
			}
			// Set cookie
			http.SetCookie(w, &http.Cookie{
				Name:  "username",
				Value: userName,
				Path:  "/", // prefix-agnostic
			})
		} else {
			// Show name entry form
			if err := templates.ExecuteTemplate(w, "enter_name.html", nil); err != nil {
				http.Error(w, "Template error", http.StatusInternalServerError)
			}
			return
		}
	}

	// User has a name now, handle dice roll
	if r.Method == http.MethodPost && r.FormValue("action") == "roll" {
		dice := rand.Intn(20) + 1
		entry := fmt.Sprintf("%s rolled a %d", userName, dice)
		room.Log = append(room.Log, entry)
	}

	templates.ExecuteTemplate(w, "room.html", room)
}

func main() {
	fmt.Println("Dice Room begins...")
	args, err := ReadArgs()
	hostPrefix = args.HostPrefix //set the hostPrefix globally.
	if err != nil {
		log.Fatal("Error parsing args: ", err)
	}

	rand.Seed(time.Now().UnixNano())

	// Serve static files
	fs := http.FileServer(http.FS(content))
	http.Handle("/static/", fs)

	// Handlers
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/room/", roomHandler)

	portStr := strconv.Itoa(args.Port)
	log.Println("Listening on " + portStr)
	fmt.Println("Dice room is ready.") //<-- Healthy Regex
	log.Fatal(http.ListenAndServe(":"+portStr, nil))
}
