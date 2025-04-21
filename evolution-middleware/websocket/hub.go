package websocket

import (
    "encoding/json"
    "sync"
    
    "github.com/gorilla/websocket"
)

// Hub mantém o controle de todas as conexões websocket ativas
type Hub struct {
    clients    map[*Client]bool
    register   chan *Client
    unregister chan *Client
    broadcast  chan []byte
    mutex      sync.Mutex
}

// Client representa uma conexão websocket com um cliente
type Client struct {
    hub      *Hub
    conn     *websocket.Conn
    send     chan []byte
    userID   string
    instance string
}

// NewHub cria um novo hub para gerenciar conexões websocket
func NewHub() *Hub {
    return &Hub{
        clients:    make(map[*Client]bool),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        broadcast:  make(chan []byte),
    }
}

// Run inicia o hub e gerencia as conexões
func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mutex.Lock()
            h.clients[client] = true
            h.mutex.Unlock()
        case client := <-h.unregister:
            h.mutex.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
            h.mutex.Unlock()
        case message := <-h.broadcast:
            h.mutex.Lock()
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
            h.mutex.Unlock()
        }
    }
}

// BroadcastToInstance envia uma mensagem para clientes inscritos em uma instância específica
func (h *Hub) BroadcastToInstance(instance string, message interface{}) {
    data, err := json.Marshal(message)
    if err != nil {
        return
    }
    
    h.mutex.Lock()
    for client := range h.clients {
        if client.instance == instance {
            client.send <- data
        }
    }
    h.mutex.Unlock()
}