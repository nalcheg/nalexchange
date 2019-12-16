package service

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/nalcheg/nalexchange/types"
)

type SaverServiceInterface interface {
	Listen() error
}

type saverService struct {
	db      *sqlx.DB
	inCh    chan types.Order
	inDeals chan types.Deal
	saveCh  chan struct{}
	orders  []types.Order
	deals   []types.Deal
}

func NewSaver(db *sqlx.DB, inCh chan types.Order, inDeals chan types.Deal) *saverService {
	return &saverService{db: db, saveCh: make(chan struct{}), inCh: inCh, inDeals: inDeals}
}

func (ss saverService) Listen() error {
	log.Print("saverService starting listening channels")
	go func() {
		for {
			ss.saveCh <- struct{}{}
			time.Sleep(3 * time.Second)
		}
	}()

	for {
		select {
		case order := <-ss.inCh:
			ss.orders = append(ss.orders, order)
		case deal := <-ss.inDeals:
			ss.deals = append(ss.deals, deal)
		case <-ss.saveCh:
			if err := ss.saveOrders(); err != nil {
				return err
			}
			if err := ss.saveDeals(); err != nil {
				return err
			}
		}
	}
}

func (ss *saverService) saveOrders() error {
	if len(ss.orders) > 0 {
		tx, err := ss.db.Begin()
		if err != nil {
			return err
		}
		stmt, err := tx.Prepare(`
			INSERT INTO exchange.orders (id,time,user_id,side,price,amount) VALUES (?,?,?,?,?,?)
		`)
		if err != nil {
			return err
		}

		for _, o := range ss.orders {
			amount, _ := o.Amount.Float64()
			if _, err := stmt.Exec(o.ID.String(), o.Time, o.UserID, o.Side, o.Price, amount); err != nil {
				return err
			}
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		ss.orders = nil
	}

	return nil
}

func (ss *saverService) saveDeals() error {
	if len(ss.deals) > 0 {
		tx, err := ss.db.Begin()
		if err != nil {
			return err
		}
		stmt, err := tx.Prepare(`
			INSERT INTO exchange.deals (id, time, left_order_id, right_order_id, price, amount) VALUES (?,?,?,?,?,?)
		`)
		if err != nil {
			return err
		}

		for _, d := range ss.deals {
			amount, _ := d.Amount.Float64()
			if _, err := stmt.Exec(d.ID, d.Time, d.LeftOrderId, d.RigthOrderId, d.Price, amount); err != nil {
				return err
			}
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		ss.deals = nil
	}

	return nil
}
