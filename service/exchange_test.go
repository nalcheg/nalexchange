package service

import (
	"log"
	"sync"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/nalcheg/nalexchange/types"
	"github.com/shopspring/decimal"
)

func BenchmarkTrade(b *testing.B) {
	gofakeit.Seed(0)
	saveOrders := make(chan types.Order)
	saveDeals := make(chan types.Deal)
	go func() {
		for {
			select {
			case <-saveOrders:
			case <-saveDeals:
			}
		}
	}()
	service := NewExchangeService("pair", nil, saveOrders, saveDeals, nil)
	for i := 0; i < b.N; i++ {
		if err := service.Add(&types.Order{
			Side:   types.Side(gofakeit.Number(0, 1)),
			Price:  gofakeit.Float64Range(0, 10),
			Amount: decimal.NewFromFloat(gofakeit.Float64Range(0, 10)),
		}); err != nil {
			b.Fatal(err)
		}
	}
}

func TestSort(t *testing.T) {
	gofakeit.Seed(0)

	saveOrders := make(chan types.Order)
	saveDeals := make(chan types.Deal)
	go func() {
		for {
			select {
			case <-saveOrders:
			case <-saveDeals:
			}
		}
	}()

	serviceForSell := NewExchangeService("pair", nil, saveOrders, saveDeals, nil)
	serviceForBuy := NewExchangeService("pair", nil, saveOrders, saveDeals, nil)

	for i := 0; i < 50; i++ {
		if err := serviceForSell.Add(&types.Order{
			Side:   types.Sell,
			Price:  gofakeit.Float64Range(0, 100),
			Amount: decimal.NewFromInt(1),
		}); err != nil {
			t.Error(err)
		}
		if err := serviceForBuy.Add(&types.Order{
			Side:   types.Buy,
			Price:  gofakeit.Float64Range(0, 100),
			Amount: decimal.NewFromInt(1),
		}); err != nil {
			t.Error(err)
		}
	}

	price := float64(0)
	for _, v := range serviceForSell.domSell {
		if price > v.Price {
			t.Error()
		}
		price = v.Price
	}

	price = float64(101)
	for _, v := range serviceForBuy.domBuy {
		if price < v.Price {
			t.Error()
		}
		price = v.Price
	}

	//log.Print("SELL")
	//for _, v := range serviceForSell.domSell {
	//	log.Print(v.Price)
	//}
	//log.Print("BUY")
	//for _, v := range serviceForBuy.domBuy {
	//	log.Print(v.Price)
	//}
}

func TestTrade(t *testing.T) {
	tests := []struct {
		name                  string
		side                  types.Side
		price                 float64
		amount                decimal.Decimal
		existedOrders         []*types.Order
		expectedBuyDomAmount  decimal.Decimal
		expectedBuyDomLen     int
		expectedSellDomAmount decimal.Decimal
		expectedSellDomLen    int
	}{
		{
			name:   "Sell equal",
			side:   types.Sell,
			price:  1,
			amount: decimal.NewFromInt(1),
			existedOrders: []*types.Order{
				{
					Side:   types.Buy,
					Price:  1,
					Amount: decimal.NewFromInt(1),
				},
			},
			expectedBuyDomAmount:  decimal.NewFromInt(0),
			expectedBuyDomLen:     0,
			expectedSellDomAmount: decimal.NewFromInt(0),
			expectedSellDomLen:    0,
		}, {
			name:   "Sell part",
			side:   types.Sell,
			price:  1,
			amount: decimal.NewFromInt(1),
			existedOrders: []*types.Order{
				{
					Side:   types.Buy,
					Price:  1,
					Amount: decimal.NewFromInt(2),
				},
			},
			expectedBuyDomAmount:  decimal.NewFromInt(1),
			expectedBuyDomLen:     1,
			expectedSellDomAmount: decimal.NewFromInt(0),
			expectedSellDomLen:    0,
		}, {
			name:   "Sell greater",
			side:   types.Sell,
			price:  1,
			amount: decimal.NewFromInt(2),
			existedOrders: []*types.Order{
				{
					Side:   types.Buy,
					Price:  1,
					Amount: decimal.NewFromInt(1),
				},
			},
			expectedBuyDomAmount:  decimal.NewFromInt(0),
			expectedBuyDomLen:     0,
			expectedSellDomAmount: decimal.NewFromInt(1),
			expectedSellDomLen:    1,
		}, {
			name:   "Sell two with one",
			side:   types.Sell,
			price:  1,
			amount: decimal.NewFromInt(2),
			existedOrders: []*types.Order{
				{
					Side:   types.Buy,
					Price:  1,
					Amount: decimal.NewFromInt(1),
				}, {
					Side:   types.Buy,
					Price:  1,
					Amount: decimal.NewFromInt(1),
				},
			},
			expectedBuyDomAmount:  decimal.NewFromInt(0),
			expectedBuyDomLen:     0,
			expectedSellDomAmount: decimal.NewFromInt(0),
			expectedSellDomLen:    0,
		}, {
			name:   "Sell one from two",
			side:   types.Sell,
			price:  1,
			amount: decimal.NewFromFloat(1),
			existedOrders: []*types.Order{
				{
					Side:   types.Buy,
					Price:  1,
					Amount: decimal.NewFromFloat(1),
				}, {
					Side:   types.Buy,
					Price:  1,
					Amount: decimal.NewFromFloat(1),
				},
			},
			expectedBuyDomAmount:  decimal.NewFromFloat(1),
			expectedBuyDomLen:     1,
			expectedSellDomAmount: decimal.NewFromFloat(0),
			expectedSellDomLen:    0,
		}, {
			name:   "Buy equal",
			side:   types.Buy,
			price:  1,
			amount: decimal.NewFromInt(1),
			existedOrders: []*types.Order{
				{
					Side:   types.Sell,
					Price:  1,
					Amount: decimal.NewFromInt(1),
				},
			},
			expectedBuyDomAmount:  decimal.NewFromInt(0),
			expectedBuyDomLen:     0,
			expectedSellDomAmount: decimal.NewFromInt(0),
			expectedSellDomLen:    0,
		}, {
			name:   "Buy part",
			side:   types.Buy,
			price:  1,
			amount: decimal.NewFromInt(1),
			existedOrders: []*types.Order{
				{
					Side:   types.Sell,
					Price:  1,
					Amount: decimal.NewFromInt(2),
				},
			},
			expectedBuyDomAmount:  decimal.NewFromInt(0),
			expectedBuyDomLen:     0,
			expectedSellDomAmount: decimal.NewFromInt(1),
			expectedSellDomLen:    1,
		}, {
			name:   "Buy greater",
			side:   types.Buy,
			price:  1,
			amount: decimal.NewFromInt(2),
			existedOrders: []*types.Order{
				{
					Side:   types.Sell,
					Price:  1,
					Amount: decimal.NewFromInt(1),
				},
			},
			expectedBuyDomAmount:  decimal.NewFromInt(1),
			expectedBuyDomLen:     1,
			expectedSellDomAmount: decimal.NewFromInt(0),
			expectedSellDomLen:    0,
		}, {
			name:   "Buy two with one",
			side:   types.Buy,
			price:  1,
			amount: decimal.NewFromInt(2),
			existedOrders: []*types.Order{
				{
					Side:   types.Sell,
					Price:  1,
					Amount: decimal.NewFromInt(1),
				}, {
					Side:   types.Sell,
					Price:  1,
					Amount: decimal.NewFromInt(1),
				},
			},
			expectedBuyDomAmount:  decimal.NewFromInt(0),
			expectedBuyDomLen:     0,
			expectedSellDomAmount: decimal.NewFromInt(0),
			expectedSellDomLen:    0,
		}, {
			name:   "Buy one from two",
			side:   types.Buy,
			price:  1,
			amount: decimal.NewFromFloat(1),
			existedOrders: []*types.Order{
				{
					Side:   types.Sell,
					Price:  1,
					Amount: decimal.NewFromFloat(1),
				}, {
					Side:   types.Sell,
					Price:  1,
					Amount: decimal.NewFromFloat(1),
				},
			},
			expectedBuyDomAmount:  decimal.NewFromFloat(0),
			expectedBuyDomLen:     0,
			expectedSellDomAmount: decimal.NewFromFloat(1),
			expectedSellDomLen:    1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			saveOrders := make(chan types.Order, 100)
			saveDeals := make(chan types.Deal, 100)
			go func() {
				for {
					select {
					case <-saveOrders:
					case <-saveDeals:
					}
				}
			}()
			service := NewExchangeService("pair", nil, saveOrders, saveDeals, nil)

			for _, eo := range test.existedOrders {
				if err := service.Add(eo); err != nil {
					t.Fatal(err)
				}
			}

			if err := service.Add(&types.Order{
				Side:   test.side,
				Price:  test.price,
				Amount: test.amount,
			}); err != nil {
				t.Fatal(err)
			}

			expectedBuyDomAmount, expectedBuyDomLen, expectedSellDomAmount, expectedSellDomLen := calcDom(service)
			if !expectedBuyDomAmount.Equal(test.expectedBuyDomAmount) {
				t.Error()
			}
			if expectedBuyDomLen != test.expectedBuyDomLen {
				t.Error()
			}
			if !expectedSellDomAmount.Equal(test.expectedSellDomAmount) {
				t.Error()
			}
			if expectedSellDomLen != test.expectedSellDomLen {
				t.Error()
			}
		})
	}
}

func TestTradeSingle(t *testing.T) {
	tests := []struct {
		name                  string
		side                  types.Side
		price                 float64
		amount                decimal.Decimal
		existedOrders         []*types.Order
		expectedBuyDomAmount  decimal.Decimal
		expectedBuyDomLen     int
		expectedSellDomAmount decimal.Decimal
		expectedSellDomLen    int
	}{
		{
			name:   "Single",
			side:   types.Sell,
			price:  1,
			amount: decimal.NewFromFloat(0.5),
			existedOrders: []*types.Order{
				{
					Side:   types.Buy,
					Price:  1,
					Amount: decimal.NewFromFloat(0.25),
				},
			},
			expectedBuyDomAmount:  decimal.NewFromFloat(0),
			expectedBuyDomLen:     0,
			expectedSellDomAmount: decimal.NewFromFloat(0.25),
			expectedSellDomLen:    1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			saveOrders := make(chan types.Order, 10)
			saveDeals := make(chan types.Deal, 10)

			service := NewExchangeService("pair", nil, saveOrders, saveDeals, nil)

			go func() {
				time.Sleep(5000 * time.Millisecond)
				close(saveOrders)
				close(saveDeals)
			}()

			for _, eo := range test.existedOrders {
				if err := service.Add(eo); err != nil {
					t.Fatal(err)
				}
			}

			if err := service.Add(&types.Order{
				Side:   test.side,
				Price:  test.price,
				Amount: test.amount,
			}); err != nil {
				t.Fatal(err)
			}

			expectedBuyDomAmount, expectedBuyDomLen, expectedSellDomAmount, expectedSellDomLen := calcDom(service)
			if !expectedBuyDomAmount.Equal(test.expectedBuyDomAmount) {
				t.Error()
			}
			if expectedBuyDomLen != test.expectedBuyDomLen {
				t.Error()
			}
			if !expectedSellDomAmount.Equal(test.expectedSellDomAmount) {
				t.Error()
			}
			if expectedSellDomLen != test.expectedSellDomLen {
				t.Error()
			}
		})
	}
}

func TestTradeSingleTemp(t *testing.T) {
	tests := []struct {
		name                  string
		side                  types.Side
		price                 float64
		amount                decimal.Decimal
		existedOrders         []types.Order
		expectedBuyDomAmount  decimal.Decimal
		expectedBuyDomLen     int
		expectedSellDomAmount decimal.Decimal
		expectedSellDomLen    int
	}{
		{
			name: "Single",
			existedOrders: []types.Order{
				{
					Side:   types.Sell,
					Price:  1.01,
					Amount: decimal.NewFromFloat(0.75),
				}, {
					Side:   types.Buy,
					Price:  1.01,
					Amount: decimal.NewFromFloat(0.25),
				}, {
					Side:   types.Buy,
					Price:  1.01,
					Amount: decimal.NewFromFloat(0.25),
				}, {
					Side:   types.Buy,
					Price:  1.01,
					Amount: decimal.NewFromFloat(0.5),
				}, {
					Side:   types.Buy,
					Price:  1.01,
					Amount: decimal.NewFromFloat(2),
				},
			},
			expectedBuyDomAmount:  decimal.NewFromFloat(2.25),
			expectedBuyDomLen:     2,
			expectedSellDomAmount: decimal.NewFromFloat(0),
			expectedSellDomLen:    0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			trCh := make(chan types.Order)
			saveOrders := make(chan types.Order, 10)
			saveDeals := make(chan types.Deal, 10)

			service := NewExchangeService("pair", trCh, saveOrders, saveDeals, nil)

			go func() {
				log.Fatal(service.Listen())
			}()

			time.Sleep(10 * time.Millisecond)
			for _, eo := range test.existedOrders {
				trCh <- eo
			}
			time.Sleep(10 * time.Millisecond)

			type decSum struct {
				sync.Mutex
				d decimal.Decimal
			}
			var ordersSum, dealsSum decSum

			go func() {
				for o := range saveOrders {
					ordersSum.Lock()
					ordersSum.d = ordersSum.d.Add(o.Amount)
					ordersSum.Unlock()
				}
			}()
			time.Sleep(10 * time.Millisecond)

			go func() {
				for d := range saveDeals {
					ordersSum.Lock()
					dealsSum.d = dealsSum.d.Add(d.Amount)
					ordersSum.Unlock()
				}
			}()
			time.Sleep(10 * time.Millisecond)

			ordersSum.Lock()
			if !ordersSum.d.Equal(decimal.NewFromFloat(3.75)) {
				t.Error()
			}

			dealsSum.Lock()
			if !dealsSum.d.Equal(decimal.NewFromFloat(0.75)) {
				t.Error()
			}
		})
	}
}

func calcDom(service *exchangeService) (decimal.Decimal, int, decimal.Decimal, int) {
	expectedSellDomAmount := decimal.NewFromInt(0)
	for _, v := range service.domSell {
		expectedSellDomAmount = expectedSellDomAmount.Add(v.Amount)
	}

	expectedBuyDomAmount := decimal.NewFromInt(0)
	for _, v := range service.domBuy {
		expectedBuyDomAmount = expectedBuyDomAmount.Add(v.Amount)
	}

	return expectedBuyDomAmount, len(service.domBuy), expectedSellDomAmount, len(service.domSell)
}
