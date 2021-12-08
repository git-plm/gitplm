package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/shanna/sexp"
)

func main() {
	sexpfile := os.Args[1]
	fmt.Println("reading: ", sexpfile)
	data, err := ioutil.ReadFile(sexpfile)
	if err != nil {
		log.Fatal("Error reading file: ", err)
	}

	sexps, err := sexp.Unmarshal(data)
	if err != nil {
		log.Fatal("Error parsing file: ", err)
	}

	fmt.Printf("exps: %T\n", sexps)
	fmt.Printf("exps[0]: %T\n", sexps[0])
	fmt.Printf("exps[0]: %v\n", string(sexps[0].([]byte)))
	fmt.Printf("exps[1]: %T\n", sexps[1])

	_ = sexps
}
