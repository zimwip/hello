package main

import (
	"fmt"
	"github.com/zimwip/hello/stringutil"
)

type PayloadCollection struct {
	WindowsVersion string    `json:"version"`
	Token          string    `json:"token"`
	Payloads       []Payload `json:"data"`
}

type Payload struct {
	// [redacted]
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
}

func multiplyByTwo(in <-chan int, out chan<- int) {
	fmt.Println("Initializing goroutine...")
	num := <-in
	result := num * 2
	out <- result
}
