package main

import (
	"dice_room/model"
	"dice_room/store"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("indexHandler: %s %s", r.Method, r.URL.String())
	if r.Method == http.MethodPost {
		r.ParseForm()
		roomName := strings.TrimSpace(r.FormValue("roomName"))
		room, err := s.store.CreateRoom(roomName)
		if err != nil {
			http.Error(w, "Could not create room", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, s.hostPrefix+"/room/"+room.ID, http.StatusSeeOther)
		return
	}

	if err := s.templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func (s *Server) roomHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("roomHandler: %s %s", r.Method, r.URL.String())
	roomID := r.URL.Path[len("/room/"):]

	room, err := s.store.GetRoom(roomID)
	if err != nil {
		if errors.Is(err, store.ErrRoomNotFound) {
			w.WriteHeader(http.StatusNotFound)
			s.templates.ExecuteTemplate(w, "not_found.html", nil)
		} else {
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
		return
	}

	userName := ""
	selectedDice := "d20"
	if cookie, err := r.Cookie("username"); err == nil {
		userName = cookie.Value
	}
	if cookie, err := r.Cookie("selectedDice"); err == nil {
		selectedDice = cookie.Value
	}

	if r.Method == http.MethodPost {
		action := r.FormValue("action")
		switch action {
		case "join":
			userName = r.FormValue("name")
			http.SetCookie(w, &http.Cookie{
				Name:  "username",
				Value: userName,
				Path:  "/room/" + roomID,
			})

		case "roll":
			desc := r.FormValue("desc")
			diceType := r.FormValue("dice")
			http.SetCookie(w, &http.Cookie{
				Name:  "selectedDice",
				Value: diceType,
				Path:  "/",
			})

			sides := 20
			if diceType != "" {
				if parsed, err := strconv.Atoi(diceType[1:]); err == nil {
					sides = parsed
				}
				selectedDice = diceType
			}

			entry := model.LogEntry{
				User:   userName,
				Dice:   diceType,
				Result: rand.Intn(sides) + 1,
				Desc:   desc,
				Time:   time.Now().Format("15:04:05"),
			}

			if err := s.store.AddEntry(roomID, entry); err != nil {
				http.Error(w, "Could not record roll", http.StatusInternalServerError)
				return
			}

			if b, err := json.Marshal(entry); err == nil {
				s.broadcaster.Send(roomID, string(b))
			}

			// Post/Redirect/Get: prevents double-roll on browser refresh.
			http.Redirect(w, r, s.hostPrefix+"/room/"+roomID, http.StatusSeeOther)
			return
		}
	}

	// Copy log under lock so we don't hold it during template rendering.
	room.Lock.Lock()
	logSnapshot := make([]model.LogEntry, len(room.Log))
	copy(logSnapshot, room.Log)
	room.Lock.Unlock()

	data := model.RoomData{
		ID:           roomID,
		RoomName:     room.RoomName,
		Log:          logSnapshot,
		UserName:     userName,
		SelectedDice: selectedDice,
		HostPrefix:   s.hostPrefix,
	}
	s.templates.ExecuteTemplate(w, "room.html", data)
}

func (s *Server) eventsHandler(w http.ResponseWriter, r *http.Request) {
	roomID := strings.TrimPrefix(r.URL.Path, "/events/")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := s.broadcaster.Subscribe(roomID)

	go func() {
		<-r.Context().Done()
		s.broadcaster.Unsubscribe(roomID, ch)
	}()

	for msg := range ch {
		fmt.Fprintf(w, "data: %s\n\n", msg)
		flusher.Flush()
	}
}
