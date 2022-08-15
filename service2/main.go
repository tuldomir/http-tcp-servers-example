package main

import (
	"errors"
	"fmt"
)

const (
	host = "localhost"
	port = "9000"

	pairsep    = "\r\n"
	eof        = pairsep + " "
	msgdelim   = ' '
	digitdelim = ","
)

// ErrNotCorrectFormat .
var ErrNotCorrectFormat = errors.New("not correct format")

func main() {

	ser, err := New(host, port)
	if err != nil {
		panic(err)
	}

	fmt.Println("server started")

	defer ser.Stop()
	ser.Run()
}
