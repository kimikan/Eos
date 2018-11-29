package helpers_test

import (
	"encoding/json"
	"fmt"
	"gamble/helpers"
	"testing"
)

func Test_getAccount(t *testing.T) {
	api := helpers.NewEos()
	resp, _ := api.GetAccount("qwerasdfvcxz")
	//fmt.Printf("%d")
	bs, _ := json.Marshal(resp)
	t.Error(string(bs))
}

func Test_deposit(t *testing.T) {
	api := helpers.NewEos()
	t.Error(helpers.EosDeposit(api, "qweasdzxcvfr", 2))
}

func Test_reveal(t *testing.T) {
	api := helpers.NewEos()
	t.Error(helpers.EosReveal(api, "qweasdzxcvfr", 50))
}

func Test_getreveal(t *testing.T) {
	api := helpers.NewEos()
	e, resp := helpers.EosGetRevealResult(api)
	if e != nil {
		t.Error(e)
		return
	}
	fmt.Println(resp)
	t.Error(resp)
}

func Test_check(t *testing.T) {
	api := helpers.NewEos()
	t.Error(helpers.EosPayAndClean(api, "qweasdzxcvfr"))
}

func Test_playdice(t *testing.T) {
	api := helpers.NewEos()
	t.Error(helpers.EosPlayDice(api, "qweasdzxcvfr", 2, 40))
}
