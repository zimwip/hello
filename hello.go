package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func sayhelloName(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()       // parse arguments, you have to call this by yourself
	fmt.Println(r.Form) // print form information in server side
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	log.Println("receive query " + r.URL.Path)
	fmt.Fprintf(w, "Hello astaxie!") // send data to client side
}

func main() {

	cluster := flag.String("cluster", "http://127.0.0.1:9021", "comma separated cluster peers")
	id := flag.Int("id", 1, "node ID")
	kvport := flag.Int("port", 9121, "key-value server port")
	join := flag.Bool("join", false, "join an existing cluster")
	flag.Parse()
	fmt.Printf("cluster: %s, id: %d, kvPort: %d, join: %t\n", *cluster, *id, *kvport, *join)

	// This is where we "make" the channel, which can be used
	// to move the `int` datatype
	out := make(chan int)
	in := make(chan int)

	// We still run this function as a goroutine, but this time,
	// the channel that we made is also provided
	go multiplyByTwo(in, out)
	go multiplyByTwo(in, out)
	go multiplyByTwo(in, out)

	// Up till this point, none of the created goroutines actually do
	// anything, since they are all waiting for the `in` channel to
	// receive some data
	in <- 1
	in <- 2
	in <- 3

	// Once any output is received on this channel, print it to the console and proceed
	fmt.Println(<-out)
	fmt.Println(<-out)
	fmt.Println(<-out)

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

	http.HandleFunc("/", sayhelloName)            // set router
	err_http := http.ListenAndServe(":9090", nil) // set listen port
	if err_http != nil {
		log.Fatal("ListenAndServe: ", err_http)
	}
}

func multiplyByTwo(in <-chan int, out chan<- int) {
	fmt.Println("Initializing goroutine...")
	num := <-in
	result := num * 2
	out <- result
}
