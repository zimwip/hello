package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	cluster := flag.String("cluster", "http://127.0.0.1:9021", "comma separated cluster peers")
	id := flag.Int("id", 1, "node ID")
	kvport := flag.Int("port", 9121, "key-value server port")
	join := flag.Bool("join", false, "join an existing cluster")
	flag.Parse()
	fmt.Printf("cluster: %s, id: %d, kvPort: %d, join: %t\n", *cluster, *id, *kvport, *join)

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

	// Création d’une variable pour l’interception du signal de fin de programme
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	// Go routine (thread parallèle) d’attente de fin du programme
	// pour l’extinction de la LED et la fermeture du port
	go func() {
		<-c
		log.Println("Stopping program")
		os.Exit(0)
	}()

	sa := new(SocketServer)

	sa.Setup(":1234")

	sa.Serve()

}
