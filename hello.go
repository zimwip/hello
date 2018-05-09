package main

import (
	"fmt"
	"github.com/zimwip/hello/stringutil"
)

func main() {
	string result = stringutil.Reverse("Hello, world. \n")
	fmt.Printf(result)
}

