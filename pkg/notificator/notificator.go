package notificator

import (
	"encoding/json"
	"math/rand"
	"sync"
	"time"

	"github.com/1F47E/go-feesh/pkg/logger"
	"github.com/gofiber/websocket/v2"
)

var log = logger.Log.WithField("scope", "notificator")

type Msg struct {
	PoolSize int `json:"pool_size"`
	TotalFee int `json:"total_fee"`
	AvgFee   int `json:"avg_fee"`
	Amount   int `json:"total_amount"`
	Weight   int `json:"total_weight"`
}

type client struct {
	isClosing bool
	mu        sync.Mutex
}

type Notificator struct {
	Register   chan *websocket.Conn
	Unregister chan *websocket.Conn
	clients    map[*websocket.Conn]*client
	broadcast  chan Msg
}

func New() *Notificator {
	Register := make(chan *websocket.Conn)
	Unregister := make(chan *websocket.Conn)
	clients := make(map[*websocket.Conn]*client)
	broadcast := make(chan Msg)
	return &Notificator{Register, Unregister, clients, broadcast}
}

func (n *Notificator) Start() {
	go n.workerWsHub()
	go n.workerWsDemo()
}

func (n *Notificator) SendWsMsg(msg Msg) {
	n.broadcast <- msg
}

func (n *Notificator) workerWsHub() {
	for {
		select {
		case connection := <-n.Register:
			n.clients[connection] = &client{}
			log.Debugf("connection registered")

		case msg := <-n.broadcast:
			log.Debugf("message received: %+v", msg)
			// Send the message to all clients
			for connection, c := range n.clients {
				go func(connection *websocket.Conn, c *client) {
					c.mu.Lock()
					defer c.mu.Unlock()
					if c.isClosing {
						return
					}
					// serialize message
					msgBytes, err := json.Marshal(msg)
					if err != nil {
						log.Errorf("error on marshal msg: %v", err)
						return
					}
					if err := connection.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
						c.isClosing = true
						log.Println("write error:", err)

						err = connection.WriteMessage(websocket.CloseMessage, []byte{})
						if err != nil {
							log.Errorf("close error: %v", err)
						}
						connection.Close()
						n.Unregister <- connection
					}
				}(connection, c)
			}

		case connection := <-n.Unregister:
			// Remove the client from the hub
			delete(n.clients, connection)

			log.Println("connection unregistered")
		}
	}
}

// demo ws msg
func (n *Notificator) workerWsDemo() {
	cnt := 0
	for {
		msg := Msg{
			PoolSize: rand.Intn(10000),
			TotalFee: rand.Intn(10000),
			AvgFee:   rand.Intn(10000),
			Amount:   rand.Intn(10000),
			Weight:   rand.Intn(10000),
		}
		n.broadcast <- msg
		log.Debugf("ws msg sent: %s", msg)
		cnt++
		time.Sleep(1 * time.Second)
	}
}
