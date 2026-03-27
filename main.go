package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	conn     *websocket.Conn
	roomID   string
	playerID string
	send     chan []byte
}

type Room struct {
	players  []*Client
	moves    map[string]string
	scores   map[string]int
	mu       sync.Mutex
	round    int
}

type Hub struct {
	rooms map[string]*Room
	mu    sync.Mutex
}

var hub = &Hub{rooms: make(map[string]*Room)}

type Message struct {
	Type     string            `json:"type"`
	Payload  map[string]string `json:"payload,omitempty"`
}

func (h *Hub) getOrCreateRoom(id string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	if r, ok := h.rooms[id]; ok {
		return r
	}
	r := &Room{
		moves:  make(map[string]string),
		scores: make(map[string]int),
		round:  1,
	}
	h.rooms[id] = r
	return r
}

func broadcast(room *Room, msg interface{}) {
	data, _ := json.Marshal(msg)
	for _, p := range room.players {
		select {
		case p.send <- data:
		default:
		}
	}
}

func writePump(c *Client) {
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

var emojiRules = map[string]map[string]string{
	"🔥": {"🌊": "🌊 douses 🔥", "🌿": "🔥 burns 🌿", "💨": "💨 fans 🔥 bigger!", "⚡": "⚡ ignites 🔥"},
	"🌊": {"🔥": "🌊 douses 🔥", "⚡": "🌊 short-circuits ⚡", "🌿": "🌿 drinks 🌊", "💨": "🌊 calms 💨"},
	"🌿": {"🔥": "🔥 burns 🌿", "🌊": "🌿 drinks 🌊", "🪨": "🌿 cracks 🪨", "💨": "💨 scatters 🌿"},
	"💨": {"🔥": "💨 fans 🔥 bigger!", "🌊": "🌊 calms 💨", "🌿": "💨 scatters 🌿", "🪨": "💨 erodes 🪨"},
	"🪨": {"🌿": "🌿 cracks 🪨", "💨": "💨 erodes 🪨", "⚡": "🪨 grounds ⚡", "🌊": "🌊 smooths 🪨"},
	"⚡": {"🌊": "🌊 short-circuits ⚡", "🪨": "🪨 grounds ⚡", "🔥": "⚡ ignites 🔥", "🌿": "⚡ zaps 🌿"},
	"🎭": {"😴": "🎭 wakes 😴", "😡": "🎭 calms 😡", "😂": "😂 upstages 🎭", "🤯": "🎭 handles 🤯"},
	"😴": {"🎭": "🎭 wakes 😴", "😡": "😴 ignores 😡", "☕": "☕ wakes 😴", "🌙": "😴 loves 🌙"},
	"😡": {"😴": "😴 ignores 😡", "🎭": "🎭 calms 😡", "😂": "😂 defuses 😡", "☕": "😡 needs ☕"},
	"😂": {"🎭": "😂 upstages 🎭", "😡": "😂 defuses 😡", "🤯": "😂 cures 🤯", "😴": "😂 wakes 😴"},
	"🤯": {"😂": "😂 cures 🤯", "🎭": "🎭 handles 🤯", "☕": "🤯 needs ☕", "😴": "😴 after 🤯"},
	"☕": {"😴": "☕ wakes 😴", "😡": "😡 needs ☕", "🤯": "🤯 needs ☕", "🌙": "☕ beats 🌙"},
	"🌙": {"😴": "😴 loves 🌙", "☕": "☕ beats 🌙", "🔥": "🌙 cools 🔥", "⚡": "🌙 calms ⚡"},
	"🎲": {},
	"🌈": {},
}

var allEmojis []string

func init() {
	for e := range emojiRules {
		allEmojis = append(allEmojis, e)
	}
}

func resolveRound(a, b string) (winner string, flavor string) {
	if a == b {
		return "", "It's a tie! Great minds think alike 🤝"
	}
	if rule, ok := emojiRules[a][b]; ok {
		return "a", rule
	}
	if rule, ok := emojiRules[b][a]; ok {
		return "b", rule
	}
	// Random tiebreak for emojis with no rule
	flavors := []string{
		"The universe is indifferent 🌌",
		"A cosmic coin flip! 🪙",
		"Nobody saw that coming 👀",
	}
	if rand.Intn(2) == 0 {
		return "a", flavors[rand.Intn(len(flavors))]
	}
	return "b", flavors[rand.Intn(len(flavors))]
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		roomID = "default"
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	room := hub.getOrCreateRoom(roomID)
	room.mu.Lock()
	if len(room.players) >= 2 {
		room.mu.Unlock()
		conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"error","payload":{"msg":"Room is full!"}}`))
		conn.Close()
		return
	}

	playerID := "A"
	if len(room.players) == 1 {
		playerID = "B"
	}

	client := &Client{conn: conn, roomID: roomID, playerID: playerID, send: make(chan []byte, 16)}
	room.players = append(room.players, client)
	playerCount := len(room.players)
	room.mu.Unlock()

	go writePump(client)

	// Send welcome
	welcome, _ := json.Marshal(map[string]interface{}{
		"type":    "welcome",
		"payload": map[string]string{"playerID": playerID, "roomID": roomID},
	})
	client.send <- welcome

	if playerCount == 2 {
		start, _ := json.Marshal(map[string]interface{}{
			"type":    "start",
			"payload": map[string]string{"msg": "Both players connected! Round 1 — choose your emoji! ⚔️"},
		})
		broadcast(room, json.RawMessage(start))
	} else {
		waiting, _ := json.Marshal(map[string]interface{}{
			"type":    "waiting",
			"payload": map[string]string{"msg": "Waiting for opponent to join..."},
		})
		client.send <- waiting
	}

	defer func() {
		conn.Close()
		room.mu.Lock()
		for i, p := range room.players {
			if p == client {
				room.players = append(room.players[:i], room.players[i+1:]...)
				break
			}
		}
		room.mu.Unlock()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var m Message
		if err := json.Unmarshal(msg, &m); err != nil {
			continue
		}
		if m.Type == "move" {
			emoji := m.Payload["emoji"]
			room.mu.Lock()
			room.moves[playerID] = emoji
			bothMoved := len(room.moves) == 2
			room.mu.Unlock()

			if bothMoved {
				room.mu.Lock()
				moveA := room.moves["A"]
				moveB := room.moves["B"]
				winner, flavor := resolveRound(moveA, moveB)
				if winner == "a" {
					room.scores["A"]++
				} else if winner == "b" {
					room.scores["B"]++
				}
				result := map[string]interface{}{
					"type": "result",
					"payload": map[string]interface{}{
						"moveA":   moveA,
						"moveB":   moveB,
						"winner":  winner,
						"flavor":  flavor,
						"scoreA":  room.scores["A"],
						"scoreB":  room.scores["B"],
						"round":   room.round,
					},
				}
				room.moves = make(map[string]string)
				room.round++
				room.mu.Unlock()
				broadcast(room, result)
			} else {
				// Notify both that one player has chosen
				waiting, _ := json.Marshal(map[string]interface{}{
					"type":    "waiting_move",
					"payload": map[string]string{"msg": "Waiting for opponent's move..."},
				})
				broadcast(room, json.RawMessage(waiting))
			}
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/ws", handleWS)

	log.Printf("🎲 Mood Duel running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
