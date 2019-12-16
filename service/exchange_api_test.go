// +build runmanual

package service

import (
	"bytes"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/shopspring/decimal"

	"github.com/nalcheg/nalexchange/types"
)

func TestExchangeService_Add_FromSlice(t *testing.T) {
	gofakeit.Seed(0)
	var os []*types.Order
	os = append(os, &types.Order{
		Side:   types.Buy,
		Price:  math.Round(gofakeit.Float64Range(0, 100)*10000000) / 10000000,
		Amount: decimal.NewFromFloat(gofakeit.Float64Range(0, 100)).Truncate(8),
	})
	os = append(os, &types.Order{
		Side:   types.Sell,
		Price:  os[0].Price,
		Amount: os[0].Amount.DivRound(decimal.NewFromInt(2), 8),
	})
	os = append(os, &types.Order{
		Side:   types.Sell,
		Price:  os[0].Price,
		Amount: os[0].Amount.DivRound(decimal.NewFromInt(2), 8),
	})

	log.Print(os[0].Amount.String())
	log.Print(os[1].Amount.String())
	log.Print(os[2].Amount.String())

	c := &http.Client{}

	for _, o := range os {
		go send(c, o)
	}

	resp, err := c.Get("http://127.0.0.1:58080/list")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Error(resp.StatusCode)
	}
}

func TestExchangeService_Add_InGoroutines(t *testing.T) {
	start := time.Now()
	c := &http.Client{}

	oneAmount := decimal.NewFromInt(1)

	for i := 0; i < 3000; i++ {
		go send(c, &types.Order{
			Side:   types.Buy,
			Price:  1.1,
			Amount: oneAmount,
		})
	}

	time.Sleep(3000 * time.Millisecond)
	if _, err := c.Get("http://127.0.0.1:58080/list"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1000 * time.Millisecond)

	for i := 0; i < 3000; i++ {
		go send(c, &types.Order{
			Side:   types.Sell,
			Price:  1.1,
			Amount: oneAmount,
		})
	}

	time.Sleep(4000 * time.Millisecond)

	if _, err := c.Get("http://127.0.0.1:58080/list"); err != nil {
		t.Fatal(err)
	}
	t.Log("elapsed ", time.Since(start))
}

func TestExchangeService_Add_ManyInGoroutines(t *testing.T) {
	c := &http.Client{}

	for i := 0; i < 8000; i++ {
		go send(c, &types.Order{
			Side:   types.Buy,
			Price:  1,
			Amount: decimal.NewFromFloat(1),
		})
	}

	time.Sleep(30000 * time.Millisecond)

	for i := 0; i < 7998; i++ {
		go send(c, &types.Order{
			Side:   types.Sell,
			Price:  1,
			Amount: decimal.NewFromFloat(1),
		})
	}

	time.Sleep(30000 * time.Millisecond)

	resp, err := c.Get("http://127.0.0.1:58080/list")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Error(resp.StatusCode)
	}
}

func send(c *http.Client, order *types.Order) {
	body, err := json.Marshal(order)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := c.Post("http://127.0.0.1:58080/trade", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatal(err)
	}
}
