package main

type Instrument struct {
	doneChan  chan struct{}
	inputChan chan input
	buyChan   chan input
	sellChan  chan input
}

func (inst *Instrument) runInstrument() {
	for {
		select {
		case in := <-inst.inputChan:
			if in.orderType == inputBuy {
				inst.buyChan <- in
			} else if in.orderType == inputSell {
				inst.sellChan <- in
			}
		}
	}
}

func makeInstrument() chan input {
	instrument := Instrument{
		doneChan:  make(chan struct{}),
		inputChan: make(chan input),
		buyChan:   make(chan input),
		sellChan:  make(chan input),
	}
	go instrument.runInstrument()
	return instrument.inputChan
}
