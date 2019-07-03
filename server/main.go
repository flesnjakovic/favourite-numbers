package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var redisPool = newPool()

func newPool() *redis.Pool {
	return &redis.Pool{
		// Maximum number of idle connections in the pool.
		MaxIdle: 80,
		// Max number of connections
		MaxActive: 12000,
		// Dial is an application supplied function for creating and
		// configuring a connection.
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "redis:6379")
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home Page")
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// Upgrade this connection to a WebSocket connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	// Subscribe to user_list channel
	go subscribe(ws, redisPool.Get())
	listener(ws, redisPool.Get())
}

func listener(wsConn *websocket.Conn, redisConn redis.Conn) {
	defer redisConn.Close()
	for {
		// Read in a message
		_, p, err := wsConn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		msg := strings.Split(string(p), " ")

		// Check if message is properly formatted
		if len(msg) == 1 && msg[0] != "ping" && msg[0] != "list" || len(msg) > 2 {
			writeToSocket(wsConn, "Bad format! Specify it as: NAME NUMBER")
			continue
		}

		// Check if second parameter can be converted to integer
		if len(msg) == 2 {
			if _, err := strconv.Atoi(msg[1]); err != nil {
				writeToSocket(wsConn, "Second parameter has to be int")
				continue
			}

			publishWork(redisConn, strings.Join(msg, " "))
		}
	}
}

func writeToSocket(ws *websocket.Conn, msg string) {
	if err := ws.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		log.Println(err)
	}
}

func publishWork(redisConn redis.Conn, msg string) {
	_, err := redisConn.Do("PUBLISH", "work", msg)
	if err != nil {
		log.Println(err)
	}
}

func subscribe(wsConn *websocket.Conn, redisConn redis.Conn) {
	pubsub := redis.PubSubConn{Conn: redisConn}

	pubsub.Subscribe("user_list")
	for redisConn.Err() == nil {
		switch msg := pubsub.Receive().(type) {
		case redis.Message:
			writeToSocket(wsConn, string(msg.Data))
		case redis.Subscription:
			writeToSocket(wsConn, fmt.Sprintf("Subscription: %s %s %d\n", msg.Kind, msg.Channel, msg.Count))
		case error:
			writeToSocket(wsConn, fmt.Sprintf("error: %v\n", msg))
		}
	}
}

func setupRoutes() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
