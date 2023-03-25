package main

import (
	"fmt"
	"time"
)

type OrderBook struct {
	inputChan   chan ClientOrder
	instmtChans map[string]chan ClientOrder
	outputChan  chan string
}

func (ob *OrderBook) ensureInstrumentExists(inst string) chan ClientOrder {
	instmtChan, ok := ob.instmtChans[inst]
	if !ok {
		instmtChan = makeInstrument(ob.outputChan, inst)
		ob.instmtChans[inst] = instmtChan
	}
	return instmtChan
}

func (ob *OrderBook) handleOrder(clientOrder ClientOrder) {
	in := clientOrder.in
	instmtChan := ob.ensureInstrumentExists(in.instrument)
	switch in.orderType {
	case inputBuy:
		instmtChan <- clientOrder
	case inputSell:
		instmtChan <- clientOrder
	case inputCancel:
		instmtChan <- clientOrder
	}
}

func (ob *OrderBook) runOrderBook() {
	for {
		select {
		case clientOrder := <-ob.inputChan:
			ob.handleOrder(clientOrder)
		case outputStr := <-ob.outputChan:
			fmt.Printf("%s %v \n", outputStr, time.Now().UnixNano())
		}
	}
}

func makeOrderBook(inputChan chan ClientOrder) {
	ob := OrderBook{
		inputChan:   inputChan,
		instmtChans: make(map[string]chan ClientOrder),
		outputChan:  make(chan string),
	}
	go ob.runOrderBook()
}
