package main

type OrderBook struct{
  instruments map[string]chan input
  input_chan chan input
}

func (ob *OrderBook) ensureInstrumentExists(inst string) {
  _, ok := ob.instruments[inst]
  if !ok {
    inst_ch := make(chan input)

    // instrument := Instrument{inst_ch}
    //go instrument.runInstrument(ob.input_chan)

    ob.instruments[inst] = inst_ch
  }
}

func (ob *OrderBook) runOrderBook() {
  for in := range ob.input_chan{
    ob.ensureInstrumentExists(in.instrument)

    inst_ch := ob.instruments[in.instrument]

    inst_ch<-in
  }
}
