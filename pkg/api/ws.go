package api

import (
	"fmt"
	"sync"
	"time"

	"github.com/1F47E/go-feesh/pkg/logger"
	"github.com/gofiber/websocket/v2"
)

// Websocket client
type client struct {
	isClosing bool
	mu        sync.Mutex
}

var log = logger.Log.WithField("scope", "api.ws.hub")

func (a *Api) SendWsMsg(msg string) {
	a.broadcast <- msg
}

func (a *Api) workerWsHub() {
	for {
		select {
		case connection := <-a.register:
			a.clients[connection] = &client{}
			log.Debugf("connection registered")

		case message := <-a.broadcast:
			log.Debugf("message received: %s", message)
			// Send the message to all clients
			for connection, c := range a.clients {
				go func(connection *websocket.Conn, c *client) {
					c.mu.Lock()
					defer c.mu.Unlock()
					if c.isClosing {
						return
					}
					if err := connection.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
						c.isClosing = true
						log.Println("write error:", err)

						err = connection.WriteMessage(websocket.CloseMessage, []byte{})
						if err != nil {
							log.Errorf("close error: %v", err)
						}
						connection.Close()
						a.unregister <- connection
					}
				}(connection, c)
			}

		case connection := <-a.unregister:
			// Remove the client from the hub
			delete(a.clients, connection)

			log.Println("connection unregistered")
		}
	}
}

// demo ws msg
func (a *Api) workerWsDemo() {
	cnt := 0
	for {
		msg := fmt.Sprintf("Hello, %d", cnt)
		a.broadcast <- msg
		log.Debugf("ws msg sent: %s", msg)
		cnt++
		time.Sleep(1 * time.Second)
	}
}
