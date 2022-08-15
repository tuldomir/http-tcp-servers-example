package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"service1/database"
	"service1/handlers"
	"service1/services"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

const (
	redishost  = "localhost"
	redisport  = "6379"
	remotehost = "localhost"
	remoteport = "9000"
)

var (
	serverhost string
	serverport string
)

func init() {
	flag.StringVar(&serverhost, "host", "localhost", "provide host")
	flag.StringVar(&serverport, "port", "8080", "provide port")
}

func main() {

	flag.Parse()

	opts := &redis.Options{
		Addr:     net.JoinHostPort(redishost, redisport),
		Password: "",
		DB:       0,
	}

	client := initRedis(opts)
	db := database.NewDB(client)
	defer func() {
		if err := db.Stop(); err != nil {
			fmt.Println(err)
		}
	}()

	connector := services.NewTCPConnector(net.JoinHostPort(remotehost, remoteport))

	serv := services.NewTService(db, connector)
	h := handlers.NewHandler(serv)

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("hello")) })
	r.HandleFunc("/test1", h.IncrementByHandler()).Methods(http.MethodPost)
	r.HandleFunc("/test2", h.HashStringHandler()).Methods(http.MethodPost)
	r.HandleFunc("/test3", h.MulStringValHandler()).Methods(http.MethodPost)

	fmt.Println("service started")
	http.ListenAndServe(net.JoinHostPort(serverhost, serverport), r)

}

func initRedis(opts *redis.Options) *redis.Client {
	client := redis.NewClient(opts)
	_, err := client.Ping(client.Context()).Result()

	if err != nil {
		panic(fmt.Errorf("cant connect to client, %w", err))
	}

	return client
}
