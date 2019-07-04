package main

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

var redisPool = newPool()

func newPool() *redis.Pool {
	return &redis.Pool{
		// Maximum number of idle connections in the pool.
		MaxIdle: 80,
		// max number of connections
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

type user struct {
	name, favouriteNumber string
}

func publishUsers(conn redis.Conn) {
	defer conn.Close()

	_, err := redisPool.Get().Do("PUBLISH", "user_list", getAllUsers(conn))
	if err != nil {
		log.Println(err)
	}
}

func getAllUsers(conn redis.Conn) string {
	list, err := redis.Strings(conn.Do("HGETALL", "users"))
	if err != nil {
		log.Println(err)
	}

	// Convert Redis result to user struct slice
	users := make([]user, len(list)/2)
	for idx := 0; idx < len(list); idx += 2 {
		users[idx/2].name = list[idx]
		users[idx/2].favouriteNumber = list[idx+1]
	}

	sort.Slice(users[:], func(i, j int) bool {
		return users[i].name < users[j].name
	})

	// Convert user list to string
	var buffer bytes.Buffer
	for _, user := range users {
		buffer.WriteString("\n")
		buffer.WriteString(user.name)
		buffer.WriteString(" ")
		buffer.WriteString(user.favouriteNumber)
	}

	return buffer.String()
}

func setFavouriteNumber(msg []string) {
	setConn := redisPool.Get()
	defer setConn.Close()

	msg[0] = strings.TrimPrefix(msg[0], "[")
	msg[1] = strings.TrimSuffix(msg[1], "]")

	i, err := strconv.Atoi(msg[1])
	if err != nil {
		log.Println(err)
	}

	_, err = setConn.Do("HSET", "users", msg[0], i)
	if err != nil {
		log.Println(err)
	}
}

func receiveWork(redisConn redis.Conn, pubsub redis.PubSubConn) {
	// While not a permanent error on the connection.
	for redisConn.Err() == nil {
		switch v := pubsub.Receive().(type) {
		case redis.Message:
			msg := strings.Split(string(v.Data), " ")

			if msg[0] == "list" {
				publishUsers(redisPool.Get())
				continue
			}

			setFavouriteNumber(msg)

			publishUsers(redisPool.Get())
		case redis.Subscription:
			fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		}
	}
}

func main() {
	c := redisPool.Get()
	defer c.Close()

	// Set up subscription
	pubsub := redis.PubSubConn{Conn: c}
	pubsub.Subscribe("work")
	receiveWork(c, pubsub)
}
