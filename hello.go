package main

import (
	"fmt"
	"github.com/mlgd/gpio"
	"github.com/zimwip/hello/stringutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
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
	fmt.Fprintf(w, "Hello astaxie!") // send data to client side
}

func main() {
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

	result := stringutil.Reverse("Hello, world")
	fmt.Printf(result)

	// Ouverture du port 23 en mode OUT
	pin, err := gpio.OpenPin(gpio.GPIO23, gpio.ModeOutput)
	if err != nil {
		fmt.Printf("Error opening pin! %s\n", err)
		return
	}

	// Création d’une variable pour l’interception du signal de fin de programme
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	// Go routine (thread parallèle) d’attente de fin du programme
	// pour l’extinction de la LED et la fermeture du port
	go func() {
		<-c
		pin.Clear()
		pin.Close()
		os.Exit(0)
	}()

	go func() {
		// Boucle infinie réalisant la tâche souhaitée
		for {
			// Allumage de la LED
			pin.Set()
			// Attente d’une seconde
			time.Sleep(time.Second)
			// Extinction de la LED
			pin.Clear()
			// Attente d’une seconde
			time.Sleep(time.Second)
		}
	}()

	http.HandleFunc("/", sayhelloName)       // set router
	err := http.ListenAndServe(":9090", nil) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func multiplyByTwo(in <-chan int, out chan<- int) {
	fmt.Println("Initializing goroutine...")
	num := <-in
	result := num * 2
	out <- result
}
