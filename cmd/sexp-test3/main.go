package main

import (
	"fmt"
	"log"
	"os"

	"github.com/chewxy/sexp"
)

func main() {
	sexpfile := os.Args[1]
	fmt.Println("reading: ", sexpfile)
	f, err := os.Open(sexpfile)
	if err != nil {
		log.Fatal("Error opening file: ", err)
	}

	sexps, err := sexp.Parse(f)
	if err != nil {
		log.Fatal("Error parsing file: ", err)
	}

	fmt.Printf("exps: %+v\n", sexps)
}
