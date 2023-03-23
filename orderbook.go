package main

type OrderBook struct {
	inputChan chan ClientOrder
	instChans map[string]chan input
}

func (ob *OrderBook) ensureInstrumentExists(inst string) chan input {
	instChan, ok := ob.instChans[inst]
	if !ok {
		instChan = makeInstrument()
		ob.instChans[inst] = instChan
	}
	return instChan
}

func (ob *OrderBook) handleOrder(in input, doneChan chan struct{}) {
	instChan := ob.ensureInstrumentExists(in.instrument)
	switch in.orderType {
	case inputBuy:
		instChan <- in
	case inputSell:
		instChan <- in
	case inputCancel:
		instChan <- in
	}
}

func (ob *OrderBook) runOrderBook() {
	for {
		select {
		case order := <-ob.inputChan:
			ob.handleOrder(order.in, order.doneChan)
		}
	}
}

func makeOrderBook(inputChan chan ClientOrder) {
	ob := OrderBook{inputChan: inputChan, instChans: make(map[string]chan input)}
	go ob.runOrderBook()
}
