package service

import (
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/nalcheg/nalexchange/types"
)

type ExchangeServiceInterface interface {
	Add(order *types.Order) error
	Listen() error
}

type exchangeService struct {
	pair         string
	trCh         chan types.Order
	saveOrdersCh chan types.Order
	saveDealsCh  chan types.Deal
	mgmtCh       chan types.MgmtMessage
	domSell      types.Orders
	domBuy       types.Orders
}

func NewExchangeService(pair string, trCh, saveOrdersCh chan types.Order, saveDealsCh chan types.Deal, mgmtCh chan types.MgmtMessage) *exchangeService {
	return &exchangeService{pair: pair, trCh: trCh, saveOrdersCh: saveOrdersCh, saveDealsCh: saveDealsCh, mgmtCh: mgmtCh}
}

func (es *exchangeService) Listen() error {
	log.Print("exchangeService starting listening channels")
	for {
		select {
		case order := <-es.trCh:
			if err := es.Add(&order); err != nil {
				return err
			}
		case mgmtID := <-es.mgmtCh:
			switch mgmtID {
			case types.List:
				//log.Print("SELL")
				//for _, v := range es.domSell {
				//	log.Print(v.Amount.String(), " amount || price ", v.Price.String())
				//}
				//log.Print("BUY")
				//for _, v := range es.domBuy {
				//	log.Print(v.Amount.String(), " amount || price ", v.Price.String())
				//}
				log.Printf("Buy len - %d ; Sell len - %d", len(es.domBuy), len(es.domSell))
			case types.Flush:
				es.domSell = nil
				es.domBuy = nil
				log.Print("Flush depth of market")
			}
		}
	}
}

func (es *exchangeService) Add(order *types.Order) error {
	order.Time = time.Now()
	order = es.preAdd(order)

	if order != nil {
		if order.Side == types.Sell {
			for key, value := range es.domSell {
				if order.Price <= value.Price {
					es.domSell = es.domSell.Insert(key, order)
					return nil
				}
			}
			es.domSell = es.domSell.Insert(len(es.domSell), order)
		} else if order.Side == types.Buy {
			for key, value := range es.domBuy {
				if order.Price >= value.Price {
					es.domBuy = es.domBuy.Insert(key, order)
					return nil
				}
			}
			es.domBuy = es.domBuy.Insert(len(es.domBuy), order)
		}
	}

	return nil
}

func (es *exchangeService) preAdd(order *types.Order) *types.Order {
	order.ID = uuid.New()
	now := time.Now()

	es.saveOrdersCh <- *order

	var keysToDelete []int
	returnOrder := true
	if order.Side == types.Sell {
		for key, value := range es.domBuy {
			if order.Price >= value.Price {
				if order.Amount.GreaterThanOrEqual(value.Amount) {
					keysToDelete = append(keysToDelete, key)
					if order.Amount.GreaterThan(value.Amount) {
						order.Amount = order.Amount.Sub(value.Amount)
						es.saveDealsCh <- types.Deal{
							ID:           uuid.New(),
							Time:         now,
							LeftOrderId:  es.domBuy[key].ID,
							RigthOrderId: order.ID,
							Price:        order.Price,
							Amount:       value.Amount,
						}
					} else {
						returnOrder = false
						es.saveDealsCh <- types.Deal{
							ID:           uuid.New(),
							Time:         now,
							LeftOrderId:  es.domBuy[key].ID,
							RigthOrderId: order.ID,
							Price:        order.Price,
							Amount:       order.Amount,
						}
						break
					}
				} else if order.Amount.LessThan(value.Amount) {
					es.domBuy[key].Amount = es.domBuy[key].Amount.Sub(order.Amount)
					returnOrder = false
					es.saveDealsCh <- types.Deal{
						ID:           uuid.New(),
						Time:         now,
						LeftOrderId:  es.domBuy[key].ID,
						RigthOrderId: order.ID,
						Price:        order.Price,
						Amount:       order.Amount,
					}
				}
			}
		}
	} else if order.Side == types.Buy {
		for key, value := range es.domSell {
			if order.Price >= value.Price {
				if order.Amount.GreaterThanOrEqual(value.Amount) {
					keysToDelete = append(keysToDelete, key)
					if order.Amount.GreaterThan(value.Amount) {
						order.Amount = order.Amount.Sub(value.Amount)
						es.saveDealsCh <- types.Deal{
							ID:           uuid.New(),
							Time:         now,
							LeftOrderId:  es.domSell[key].ID,
							RigthOrderId: order.ID,
							Price:        order.Price,
							Amount:       value.Amount,
						}
					} else {
						returnOrder = false
						es.saveDealsCh <- types.Deal{
							ID:           uuid.New(),
							Time:         now,
							LeftOrderId:  es.domSell[key].ID,
							RigthOrderId: order.ID,
							Price:        order.Price,
							Amount:       order.Amount,
						}
						break
					}
				} else if order.Amount.LessThan(value.Amount) {
					es.domSell[key].Amount = es.domSell[key].Amount.Sub(order.Amount)
					returnOrder = false
					es.saveDealsCh <- types.Deal{
						ID:           uuid.New(),
						Time:         now,
						LeftOrderId:  es.domSell[key].ID,
						RigthOrderId: order.ID,
						Price:        order.Price,
						Amount:       order.Amount,
					}
				}
			}
		}
	}
	if order.Side == types.Sell {
		for _, key := range keysToDelete {
			es.domBuy = es.domBuy.Delete(key)
		}
	} else if order.Side == types.Buy {
		for _, key := range keysToDelete {
			es.domSell = es.domSell.Delete(key)
		}
	}

	if returnOrder {
		return order
	}

	return nil
}
