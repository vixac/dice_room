package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type subscriber chan string

var (
	subscribers = map[string][]subscriber{} // roomID → list of channels
	subMu       sync.Mutex
)

func eventsHandler(w http.ResponseWriter, r *http.Request) {
	roomID := strings.TrimPrefix(r.URL.Path, "/events/")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(subscriber, 10)

	subMu.Lock()
	subscribers[roomID] = append(subscribers[roomID], ch)
	subMu.Unlock()

	// Stream messages
	for msg := range ch {
		fmt.Fprintf(w, "data: %s\n\n", msg)
		flusher.Flush()
	}
}

// --- Embed assets ---
//
//go:embed templates/*
//go:embed static/*
var content embed.FS

var templates = template.Must(template.New("").Funcs(template.FuncMap{
	"static": func(path string) string {
		// Always resolve to absolute `/static/...`
		return hostPrefix + "/static/" + path
	},
}).ParseFS(content, "templates/*.html"))

// --- In-memory state ---
type Room struct {
	ID   string
	Log  []string
	Lock sync.Mutex
}

type RoomData struct {
	ID       string
	Log      []string
	UserName string
}

type NotFoundData struct {
	Home string
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
		log.Print("Invalid or missing room: " + roomID)
		w.WriteHeader(http.StatusNotFound)
		templates.ExecuteTemplate(w, "not_found.html", nil)
		return
	}
	room.Lock.Lock()
	defer room.Lock.Unlock()

	var userName string = ""
	if r.Method == http.MethodPost {
		action := r.FormValue("action")
		fmt.Println("action is " + action)
		switch action {
		case "join":
			userName = r.FormValue("name")

			// store in cookie or session
			http.SetCookie(w, &http.Cookie{
				Name:  "username",
				Value: userName,
				Path:  "/",
			})
		case "roll":
			userCookie, err := r.Cookie("username")
			if err != nil {
				http.Error(w, "You must join first", http.StatusForbidden)
				return
			}
			userName = userCookie.Value

			desc := r.FormValue("desc")
			dice := rand.Intn(20) + 1
			entry := fmt.Sprintf("[%s] %s rolled a %d %s",
				time.Now().Format("15:04:05"),
				userName,
				dice,
				desc,
			)
			room.Log = append(room.Log, entry)
			subMu.Lock()
			for _, ch := range subscribers[roomID] {
				select {
				case ch <- entry:
				default:
					// if a subscriber is stuck, skip
				}
			}
			subMu.Unlock()
		}
	} else {
		fmt.Println("This was not a post.")
	}

	roomData := RoomData{ID: roomID, Log: room.Log, UserName: userName}

	templates.ExecuteTemplate(w, "room.html", roomData)
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
	http.HandleFunc("/events/", eventsHandler)

	portStr := strconv.Itoa(args.Port)
	log.Println("Listening on " + portStr)
	fmt.Println("Dice room is ready.") //<-- Healthy Regex
	log.Fatal(http.ListenAndServe(":"+portStr, nil))
}
