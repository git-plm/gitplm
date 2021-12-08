package main

import (
	"fmt"
	"log"
	"os"

	"github.com/SteelSeries/bufrr"
	"github.com/nsf/sexp"
)

func main() {
	sexpfile := os.Args[1]
	fmt.Println("reading: ", sexpfile)
	f, err := os.Open(sexpfile)
	if err != nil {
		log.Fatal("Error opening file: ", err)
	}

	fRR := bufrr.NewReader(f)

	ast, err := sexp.Parse(fRR, nil)
	if err != nil {
		log.Fatal("Error parsing file: ", err)
	}

	var s interface{}

	err = ast.Unmarshal(&s)
	if err != nil {
		log.Fatal("Error unmarshalling: ", err)
	}

}
