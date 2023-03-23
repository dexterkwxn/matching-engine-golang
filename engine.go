package main

import "C"
import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
)

type ClientOrder struct {
	in       input
	doneChan chan struct{}
}

type Engine struct {
	inputChan chan ClientOrder
}

func (e *Engine) accept(ctx context.Context, conn net.Conn) {
	go func() {
		<-ctx.Done()
		conn.Close()
	}()
	go handleConn(conn, e.inputChan)
}

func handleConn(conn net.Conn, inputChan chan ClientOrder) {
	doneChan := make(chan struct{})

	defer conn.Close()
	for {
		in, err := readInput(conn)
		if err != nil {
			if err != io.EOF {
				_, _ = fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			}
			return
		}
		inputChan <- ClientOrder{in: in, doneChan: doneChan}
		<-doneChan
	}
}

func makeEngine() Engine {
	e := Engine{
		inputChan: make(chan ClientOrder),
	}
	makeOrderBook(e.inputChan)
	return e
}
