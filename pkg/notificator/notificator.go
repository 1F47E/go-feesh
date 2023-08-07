package notificator

import (
	"encoding/json"
	"sync"

	"github.com/1F47E/go-feesh/pkg/logger"
	"github.com/gofiber/websocket/v2"
)

var log = logger.Log.WithField("scope", "notificator")

type Msg struct {
	Height          int      `json:"height"`
	PoolSize        int      `json:"size"`
	PoolSizeHistory [20]uint `json:"size_history"`
	TotalFee        int      `json:"fee"`
	AvgFee          int      `json:"avg_fee"`
	Amount          int      `json:"amount"`
	Weight          int      `json:"weight"`
	FeeBuckets      [24]uint `json:"fee_buckets"`
}

type client struct {
	isClosing bool
	mu        sync.Mutex
}

type Notificator struct {
	RegisterCh         chan *websocket.Conn
	UnregisterCh       chan *websocket.Conn
	clients            map[*websocket.Conn]*client
	broadcastCh        chan Msg
	lastBroadcastedMsg Msg
}

func New(notificationsCh chan Msg) *Notificator {
	return &Notificator{
		RegisterCh:   make(chan *websocket.Conn),
		UnregisterCh: make(chan *websocket.Conn),
		clients:      make(map[*websocket.Conn]*client),
		broadcastCh:  notificationsCh,
	}
}

func (n *Notificator) Start() {
	go n.workerWsHub()
	// go n.workerWsDemo()
}

func (n *Notificator) Send(msg Msg) {
	n.broadcastCh <- msg
}

func (n *Notificator) workerWsHub() {
	for {
		select {
		case connection := <-n.RegisterCh:
			n.clients[connection] = &client{}
			log.Debugf("connection registered")

		case msg := <-n.broadcastCh:
			// avoid sending the same message
			if msg == n.lastBroadcastedMsg {
				continue
			}
			n.lastBroadcastedMsg = msg
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
						n.UnregisterCh <- connection
					}
				}(connection, c)
			}

		case connection := <-n.UnregisterCh:
			// Remove the client from the hub
			delete(n.clients, connection)

			log.Println("connection unregistered")
		}
	}
}

// demo ws msg
// func (n *Notificator) workerWsDemo() {
// 	cnt := 0
// 	for {
// 		msg := Msg{
// 			PoolSize: rand.Intn(10000),
// 			TotalFee: rand.Intn(10000),
// 			AvgFee:   rand.Intn(10000),
// 			Amount:   rand.Intn(10000),
// 			Weight:   rand.Intn(10000),
// 		}
// 		n.broadcast <- msg
// 		log.Debugf("ws msg sent: %+v", msg)
// 		cnt++
// 		time.Sleep(1 * time.Second)
// 	}
// }
