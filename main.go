package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/zimwip/hello/config"
	"github.com/zimwip/hello/router"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	cluster := flag.String("cluster", "http://127.0.0.1:9021", "comma separated cluster peers")
	id := flag.Int("id", 1, "node ID")
	kvport := flag.Int("port", 9121, "key-value server port")
	join := flag.Bool("join", false, "join an existing cluster")
	flag.Parse()
	fmt.Printf("cluster: %s, id: %d, kvPort: %d, join: %t\n", *cluster, *id, *kvport, *join)
	fmt.Println(config.GetString("app.value"))

	//create your file with desired read/write permissions
	f, err := os.OpenFile("hello.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	//defer to close when you're done with it, not because you think it's idiomatic!
	defer f.Close()
	//set output of logs to f
	log.SetOutput(f)
	//test case
	log.Println("check to make sure it works")

	sa := new(SocketServer)
	sa.Setup(":1234")
	go sa.Serve()

	srv := router.NewHttpAPI("0.0.0.0:9090")

	// Création d’une variable pour l’interception du signal de fin de programme
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	// Block until we receive our signal.
	<-c
	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}
