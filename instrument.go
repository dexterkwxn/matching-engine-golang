package main

import "container/heap"

type Order struct {
	id          uint32
	count       uint32
	price       uint32
	executionId uint32
	isSell      bool
  isCancel    bool
}

type Instrument struct {
	name          string
	inputChan     chan ClientOrder
	outputChan    chan string
	buyMatchChan  chan Order
	buyAddChan    chan Order
	sellMatchChan chan Order
	sellAddChan   chan Order
	reconChan     chan Order
	doneChans     map[uint32]chan struct{}
  orderMap      map[uint32]int // 0 = buy, 1 = sell
}

func (instmt *Instrument) runInstrument() {
	go instmt.runSellWorker()
	go instmt.runBuyWorker()
	go instmt.runReconWorker()

	for {
		select {
		case clientOrder := <-instmt.inputChan:
			in := clientOrder.in
			instmt.doneChans[in.orderId] = clientOrder.doneChan
			if in.orderType == inputBuy {
        instmt.orderMap[in.orderId] = 0
				instmt.sellMatchChan <- Order{id: in.orderId, count: in.count, price: in.price, executionId: 0, isSell: false}
			} else if in.orderType == inputSell {
        instmt.orderMap[in.orderId] = 1
				instmt.buyMatchChan <- Order{id: in.orderId, count: in.count, price: in.price, executionId: 0, isSell: true}
			} else if in.orderType == inputCancel {
        if (instmt.orderMap[in.orderId] == 0) {
          // idk if we need to change isSell to OrderType instead to detect cancel order 
          // maybe this will work?
          instmt.sellMatchChan <- Order{id: in.orderId, isCancel: true} 
        } else {
          instmt.buyMatchChan <- Order{id: in.orderId, isCancel: true}
        }
      }
		}
	}
}

func (instmt *Instrument) runSellWorker() {
	restingSells := make(PriorityQueue, 0)
	orderAddTime := 0
	for {
		select {
		case activeOrder := <-instmt.sellMatchChan:
      if activeOrder.isCancel {
        // not sure how to find the order and delete from priority queue
        continue
      }
			for {
				if activeOrder.count == 0 || len(restingSells) == 0 || restingSells[0].value.price > activeOrder.price {
					break
				}
				matchedOrder := &restingSells[0].value
				matchedCount := min(matchedOrder.count, activeOrder.count)
				matchedOrder.count -= matchedCount
				activeOrder.count -= matchedCount
				instmt.outputChan <- fmtOrderExecuted(matchedOrder.id, activeOrder.id, matchedOrder.executionId, matchedOrder.price, matchedCount)
				matchedOrder.executionId += 1
				if matchedOrder.count == 0 {
					heap.Pop(&restingSells)
				}
				if activeOrder.count == 0 {
					instmt.doneChans[activeOrder.id] <- struct{}{}
					delete(instmt.doneChans, activeOrder.id)
				}
			}
			if activeOrder.count > 0 {
				instmt.reconChan <- activeOrder
			}
		case orderToAdd := <-instmt.sellAddChan:
			item := &Item{value: orderToAdd, priority: Priority{price: -orderToAdd.price, time: -orderAddTime}}
			heap.Push(&restingSells, item)
			orderAddTime += 1

			instmt.doneChans[orderToAdd.id] <- struct{}{}
			delete(instmt.doneChans, orderToAdd.id)

			instmt.outputChan <- fmtOrderAdded(orderToAdd, instmt.name)
		}
	}
}

func (instmt *Instrument) runBuyWorker() {
	restingBuys := make(PriorityQueue, 0)
	orderAddTime := 0
	for {
		select {
		case activeOrder := <-instmt.buyMatchChan:
			for {
				if activeOrder.count == 0 || len(restingBuys) == 0 || restingBuys[0].value.price < activeOrder.price {
					break
				}
				matchedOrder := &restingBuys[0].value
				matchedCount := min(matchedOrder.count, activeOrder.count)
				matchedOrder.count -= matchedCount
				activeOrder.count -= matchedCount
				instmt.outputChan <- fmtOrderExecuted(matchedOrder.id, activeOrder.id, matchedOrder.executionId, matchedOrder.price, matchedCount)
				matchedOrder.executionId += 1
				if matchedOrder.count == 0 {
					heap.Pop(&restingBuys)
				}
				if activeOrder.count == 0 {
					instmt.doneChans[activeOrder.id] <- struct{}{}
					delete(instmt.doneChans, activeOrder.id)
				}
			}
			if activeOrder.count > 0 {
				instmt.reconChan <- activeOrder
			}
		case orderToAdd := <-instmt.buyAddChan:
			item := &Item{value: orderToAdd, priority: Priority{price: orderToAdd.price, time: orderAddTime}}
			heap.Push(&restingBuys, item)
			orderAddTime += 1

			instmt.doneChans[orderToAdd.id] <- struct{}{}
			delete(instmt.doneChans, orderToAdd.id)

			instmt.outputChan <- fmtOrderAdded(orderToAdd, instmt.name)
		}
	}
}

func (instmt *Instrument) runReconWorker() {
	prevAddedWasSell := false
	for {
		select {
		case order := <-instmt.reconChan:
			if order.isSell {
				if prevAddedWasSell {
					instmt.sellAddChan <- order
				} else {
					instmt.buyMatchChan <- order
				}
				prevAddedWasSell = true
			} else {
				if prevAddedWasSell {
					instmt.sellMatchChan <- order
				} else {
					instmt.buyAddChan <- order
				}
				prevAddedWasSell = false
			}
		}
	}
}

func makeInstrument(outputChan chan string, name string) chan ClientOrder {
	instrument := Instrument{
		name:          name,
		inputChan:     make(chan ClientOrder),
		buyMatchChan:  make(chan Order, 1),
		buyAddChan:    make(chan Order, 1),
		sellMatchChan: make(chan Order, 1),
		sellAddChan:   make(chan Order, 1),
		reconChan:     make(chan Order, 2),
		outputChan:    outputChan,
		doneChans:     make(map[uint32]chan struct{}),
    orderMap:      make(map[uint32]int),
	}
	go instrument.runInstrument()
	return instrument.inputChan
}
