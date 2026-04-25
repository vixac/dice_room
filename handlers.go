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
		http.Redirect(w, r, s.prefixFor(r)+"/room/"+room.Id, http.StatusSeeOther)
		return
	}

	if err := s.templates.ExecuteTemplate(w, "index.html", model.PageData{HostPrefix: s.prefixFor(r)}); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func (s *Server) roomHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("roomHandler: %s %s", r.Method, r.URL.String())
	roomID := r.URL.Path[len("/room/"):]

	room, err := s.store.GetRoom(roomID)
	if err != nil {
		fmt.Printf("VX: Error is %s\n", err.Error())
		if errors.Is(err, store.ErrRoomNotFound) {
			w.WriteHeader(http.StatusNotFound)
			s.templates.ExecuteTemplate(w, "not_found.html", model.PageData{HostPrefix: s.prefixFor(r)})
		} else {
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
		return
	}

	userName := ""
	if cookie, err := r.Cookie("username"); err == nil {
		userName = cookie.Value
	}

	if r.Method == http.MethodPost {
		action := r.FormValue("action")
		switch action {
		case "join":
			userName = r.FormValue("name")
			http.SetCookie(w, &http.Cookie{
				Name:     "username",
				Value:    userName,
				Path:     s.prefixFor(r) + "/room/" + roomID,
				HttpOnly: true,
				Secure:   s.secureCookies,
				SameSite: http.SameSiteLaxMode,
			})

		case "roll":
			desc := r.FormValue("desc")
			diceType := r.FormValue("dice")

			sides := 20
			if diceType != "" {
				if parsed, err := strconv.Atoi(diceType[1:]); err == nil {
					sides = parsed
				}
			}

			now := time.Now()
			entry := model.LogEntry{
				User:       userName,
				Dice:       diceType,
				Result:     rand.Intn(sides) + 1,
				Desc:       desc,
				Time:       now.Format("15:04:05"),
				UnixMillis: now.UnixMilli(),
			}

			if err := s.store.AddEntry(roomID, entry); err != nil {
				http.Error(w, "Could not record roll", http.StatusInternalServerError)
				return
			}

			if b, err := json.Marshal(entry); err == nil {
				s.broadcaster.Send(roomID, string(b))
			}

			redirect := s.prefixFor(r) + "/room/" + roomID
			fmt.Printf(" redirecting to %s\n", redirect)
			// Post/Redirect/Get: prevents double-roll on browser refresh.
			http.Redirect(w, r, redirect, http.StatusSeeOther)
			return
		}
	}

	// Copy log under lock so we don't hold it during template rendering.
	room.Lock.Lock()
	logSnapshot := make([]model.LogEntry, len(room.Log))
	copy(logSnapshot, room.Log)
	room.Lock.Unlock()

	data := model.RoomData{
		PageData: model.PageData{HostPrefix: s.prefixFor(r)},
		ID:       roomID,
		RoomName: room.RoomName,
		Log:      logSnapshot,
		UserName: userName,
	}
	s.templates.ExecuteTemplate(w, "room.html", data)
}

func (s *Server) privacyHandler(w http.ResponseWriter, r *http.Request) {
	if err := s.templates.ExecuteTemplate(w, "privacy.html", model.PageData{HostPrefix: s.prefixFor(r)}); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func (s *Server) termsHandler(w http.ResponseWriter, r *http.Request) {
	if err := s.templates.ExecuteTemplate(w, "terms.html", model.PageData{HostPrefix: s.prefixFor(r)}); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func (s *Server) contactHandler(w http.ResponseWriter, r *http.Request) {
	if err := s.templates.ExecuteTemplate(w, "contact.html", model.PageData{HostPrefix: s.prefixFor(r)}); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
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
