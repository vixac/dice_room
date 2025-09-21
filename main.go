package main

import (
	"embed"
	"encoding/json"
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

	ch := make(chan string, 16)

	subMu.Lock()
	subscribers[roomID] = append(subscribers[roomID], ch)
	subMu.Unlock()

	// remove subscriber on disconnect
	go func() {
		<-r.Context().Done()
		subMu.Lock()
		defer subMu.Unlock()
		arr := subscribers[roomID]
		for i := range arr {
			if arr[i] == ch {
				subscribers[roomID] = append(arr[:i], arr[i+1:]...)
				break
			}
		}
		close(ch)
	}()

	// send any missed messages? (optional)
	// Stream new messages from the channel
	for m := range ch {
		fmt.Fprintf(w, "data: %s\n\n", m) // m is JSON
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
	"safeHTML": func(s string) template.HTML {
		// Allow trusted HTML into template (e.g. formatted log entries)
		return template.HTML(s)
	},
}).ParseFS(content, "templates/*.html"))

// --- In-memory state ---
type Room struct {
	ID   string
	Log  []LogEntry
	Lock sync.Mutex
}

type RoomData struct {
	ID           string
	Log          []LogEntry
	UserName     string
	SelectedDice string
	HostPrefix   string
}

type NotFoundData struct {
	Home string
}

var (
	rooms      = map[string]*Room{}
	mu         sync.Mutex
	hostPrefix string
)

type LogEntry struct {
	User   string `json:"user"`
	Dice   string `json:"dice"`
	Result int    `json:"result"`
	Desc   string `json:"desc,omitempty"`
	Time   string `json:"time"`
}

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
	var selectedDice = "d20"
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
			diceType := r.FormValue("dice")

			sides := 20 // default
			if diceType != "" {
				if parsed, err := strconv.Atoi(diceType[1:]); err == nil { // strip "d"
					sides = parsed
				}
				selectedDice = diceType
			}
			dice := rand.Intn(sides) + 1

			entry := LogEntry{
				User:   userName,
				Dice:   diceType,
				Result: dice,
				Desc:   desc, // user text; we'll not treat it as HTML
				Time:   time.Now().Format("15:04:05"),
			}
			room.Log = append(room.Log, entry)
			b, err := json.Marshal(entry)
			if err == nil {
				msg := string(b)
				subMu.Lock()
				for _, ch := range subscribers[roomID] {
					select {
					case ch <- msg:
					default:
						// if a subscriber is stuck, skip
					}
				}
				subMu.Unlock()
			}

		}
	} else {
		fmt.Println("This was not a post.")
	}

	roomData := RoomData{ID: roomID, Log: room.Log, UserName: userName, SelectedDice: selectedDice, HostPrefix: hostPrefix}

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
