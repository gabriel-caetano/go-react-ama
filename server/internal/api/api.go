package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"

	"github.com/gabriel-caetano/go-react-ama/server/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type apiHandler struct {
	q           *pgstore.Queries
	r           *chi.Mux
	upgrader    websocket.Upgrader
	subscribers map[string]map[*websocket.Conn]context.CancelFunc
	mu          *sync.Mutex
}

func (h apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}

func NewHandler(q *pgstore.Queries) http.Handler {
	a := apiHandler{
		q:           q,
		upgrader:    websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		subscribers: make(map[string]map[*websocket.Conn]context.CancelFunc),
		mu:          &sync.Mutex{},
	}

	r := chi.NewRouter()
	corsMiddleware := cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	})
	r.Use(middleware.RequestID, middleware.Recoverer, middleware.Logger, corsMiddleware)

	r.Get("/subscribe/{room_id}", a.handleSubscribe)
	r.Route("/api", func(r chi.Router) {
		r.Route("/rooms", func(r chi.Router) {
			r.Post("/", a.handleCreateRoom)
			r.Get("/", a.handleGetRooms)

			r.Route("/{room_id}/messages", func(r chi.Router) {
				r.Post("/", a.handleCreateRoomMessage)
				r.Get("/", a.handleGetRoomMessages)
				r.Route("/{message_id}", func(r chi.Router) {
					r.Get("/", a.handleGetRoomMessage)
					r.Patch("/react", a.handleReactToMessage)
					r.Delete("/react", a.handleRemoveReactFromMessage)
					r.Patch("/answered", a.handleMarkMessageAsAnswered)
				})
			})
		})
	})

	a.r = r
	return a
}

const (
	MessageKindMessageCreated = "message_created"
)

type MessageMessageCreated struct {
	ID            string `json:"id"`
	Message       string `json:"message"`
	ReactionCount int64  `json:"reaction_count"`
	Answered      bool   `json:"answered"`
}

type Message struct {
	Kind   string `json:"kind"`
	Value  any    `json:"value"`
	RoomID string `json:"-"` //don't encode
}

func (h apiHandler) notifyClient(msg Message) {
	h.mu.Lock()
	defer h.mu.Unlock()

	subscribers, ok := h.subscribers[msg.RoomID]
	if !ok || len(subscribers) == 0 {
		return
	}

	for conn, cancel := range subscribers {
		if err := conn.WriteJSON(msg); err != nil {
			slog.Error("fail to send message to client", "error", err)
			cancel()
		}
	}
}

func (h apiHandler) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	RoomID, err := apiHandler.getSafeID(h, w, r, "room_id")
	if err != nil {
		panic(err)
	}

	c, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("failed to upgrade connection", "error", err)
		http.Error(w, "failed to upgrade to ws connection", http.StatusBadRequest)
		return
	}
	defer c.Close()
	ctx, cancel := context.WithCancel(r.Context())
	h.mu.Lock()

	if _, ok := h.subscribers[RoomID.rawID]; !ok {
		h.subscribers[RoomID.rawID] = make(map[*websocket.Conn]context.CancelFunc)
	}
	slog.Info("new client connected", "room_id", RoomID.rawID, "client_ip", r.RemoteAddr)
	h.subscribers[RoomID.rawID][c] = cancel
	h.mu.Unlock()

	<-ctx.Done()

	h.mu.Lock()
	delete(h.subscribers[RoomID.rawID], c)
	h.mu.Unlock()
}

func (h apiHandler) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	type _body struct {
		Theme string `json:"theme"`
	}
	var body _body
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	slog.Info("theme that should be created", "theme", body.Theme)

	roomID, err := h.q.InsertRoom(r.Context(), body.Theme)
	if err != nil {
		slog.Error("failed too insert room", "error", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	type response struct {
		ID string `json:"id"`
	}
	data, _ := json.Marshal(response{ID: roomID.String()})
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

func (h apiHandler) handleGetRooms(w http.ResponseWriter, r *http.Request) {}

func (h apiHandler) handleCreateRoomMessage(w http.ResponseWriter, r *http.Request) {
	SafeRoomID, err := apiHandler.getSafeID(h, w, r, "room_id")
	if err != nil {
		panic(err)
	}
	_, err = h.q.GetRoom(r.Context(), SafeRoomID.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "room not found", http.StatusBadRequest)
			return
		}
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	type _body struct {
		Message string `json:"message"`
	}
	var body _body
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	messageID, err := h.q.InsertMessage(
		r.Context(),
		pgstore.InsertMessageParams{RoomID: SafeRoomID.ID, Message: body.Message},
	)
	if err != nil {
		slog.Error("failed to insert message", "error", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}

	type response struct {
		ID string `json:"id"`
	}

	data, _ := json.Marshal(response{ID: messageID.String()})
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)

	go h.notifyClient(Message{
		Kind:   MessageKindMessageCreated,
		RoomID: SafeRoomID.rawID,
		Value: MessageMessageCreated{
			ID:            messageID.String(),
			Message:       body.Message,
			ReactionCount: 0,
			Answered:      false,
		},
	})
}

func (h apiHandler) handleGetRoomMessages(w http.ResponseWriter, r *http.Request) {
	SafeRoomID, err := apiHandler.getSafeID(h, w, r, "room_id")
	if err != nil {
		panic(err)
	}
	_, err = h.q.GetRoom(r.Context(), SafeRoomID.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "room not found", http.StatusBadRequest)
			return
		}
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	messages, err := h.q.GetRoomMessages(
		r.Context(), SafeRoomID.ID,
	)
	if err != nil {
		slog.Error("failed to load messages", "error", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}

	type responseMessage struct {
		ID            uuid.UUID `json:"id"`
		RoomID        uuid.UUID `json:"room_id"`
		Message       string    `json:"message"`
		ReactionCount int64     `json:"reaction_count"`
		Answered      bool      `json:"answered"`
	}

	res := []responseMessage{}
	for _, m := range messages {
		res = append(res, responseMessage{
			ID:            m.ID,
			RoomID:        m.RoomID,
			Message:       m.Message,
			ReactionCount: m.ReactionCount,
			Answered:      m.Answered,
		})
	}

	data, _ := json.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

func (h apiHandler) handleGetRoomMessage(w http.ResponseWriter, r *http.Request) {}

func (h apiHandler) handleReactToMessage(w http.ResponseWriter, r *http.Request) {
	SafeMessageID, err := apiHandler.getSafeID(h, w, r, "message_id")
	if err != nil {
		panic(err)
	}

	_, err = h.q.GetMessage(r.Context(), SafeMessageID.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "room not found", http.StatusBadRequest)
			return
		}
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	rc, err := h.q.ReactToMessage(r.Context(), SafeMessageID.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Error("failed to load message", "error", err)
			http.Error(w, "message not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to load messages", "error", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	type response struct {
		ReactionCount string `json:"reaction_count"`
	}

	data, _ := json.Marshal(response{ReactionCount: strconv.Itoa(int(rc))})
	fmt.Println(data)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

func (h apiHandler) handleRemoveReactFromMessage(w http.ResponseWriter, r *http.Request) {
	SafeMessageID, err := apiHandler.getSafeID(h, w, r, "message_id")
	if err != nil {
		panic(err)
	}
	_, err = h.q.GetMessage(r.Context(), SafeMessageID.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "room not found", http.StatusBadRequest)
			return
		}
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	rc, err := h.q.RemoveReactionFromMessage(r.Context(), SafeMessageID.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Error("failed to load message", "error", err)
			http.Error(w, "message not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to load messages", "error", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	type response struct {
		ReactionCount string `json:"reaction_count"`
	}

	data, _ := json.Marshal(response{ReactionCount: strconv.Itoa(int(rc))})
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

func (h apiHandler) handleMarkMessageAsAnswered(w http.ResponseWriter, r *http.Request) {}

type SafeID struct {
	ID    uuid.UUID
	rawID string
}

func (h apiHandler) getSafeID(w http.ResponseWriter, r *http.Request, paramID string) (SafeID, error) {

	rawID := chi.URLParam(r, paramID)
	ID, err := uuid.Parse(rawID)
	if err != nil {
		http.Error(w, "invalid room id", http.StatusBadRequest)
		return SafeID{}, err
	}

	return SafeID{
		ID:    ID,
		rawID: rawID,
	}, nil
}
