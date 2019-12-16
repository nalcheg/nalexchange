package main

import (
	"encoding/json"
	"log"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	"github.com/nalcheg/nalexchange/db"
	"github.com/nalcheg/nalexchange/service"
	"github.com/nalcheg/nalexchange/types"
)

func main() {
	tradeChannel := make(chan types.Order)
	mgmtChannel := make(chan types.MgmtMessage)
	saveOrders := make(chan types.Order)
	saveDeals := make(chan types.Deal)
	s := service.NewExchangeService("BZ/TC", tradeChannel, saveOrders, saveDeals, mgmtChannel)

	r := router.New()
	r.POST("/trade", func(ctx *fasthttp.RequestCtx) {
		var o *types.Order
		if err := json.Unmarshal(ctx.Request.Body(), &o); err != nil {
			panic(err)
		}
		tradeChannel <- *o
	})
	r.GET("/list", func(ctx *fasthttp.RequestCtx) {
		mgmtChannel <- types.List
	})
	r.GET("/flush", func(ctx *fasthttp.RequestCtx) {
		mgmtChannel <- types.Flush
	})

	go func() {
		atAddr := ":58080"
		log.Print("HTTP handler starting listening at ", atAddr)
		log.Fatal(fasthttp.ListenAndServe(atAddr, r.Handler))
	}()

	//chDB, err := db.ConnectClickhouse("tcp://127.0.0.1:9000?debug=true")
	chDB, err := db.ConnectClickhouse("tcp://127.0.0.1:9000")
	if err != nil {
		log.Fatal(err)
	}
	if err := chDB.Ping(); err != nil {
		log.Fatal(err)
	}

	saverService := service.NewSaver(chDB, saveOrders, saveDeals)
	go func() {
		log.Fatal(saverService.Listen())
	}()

	log.Fatal(s.Listen())
}
