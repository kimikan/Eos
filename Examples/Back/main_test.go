package main_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/eoscanada/eos-go"
)

const (
	kConstractAccount = "qwerasdfvcxz"
)

func NewDeposit(player eos.AccountName, eosQuantity uint64) *eos.Action {
	a := &eos.Action{
		Account: kConstractAccount,
		Name:    "deposit",
		Authorization: []eos.PermissionLevel{
			{Actor: kConstractAccount, Permission: ("active")},
		},
		ActionData: eos.NewActionData(Deposit{
			Player:   player,
			Quantity: eos.NewEOSAsset(int64(eosQuantity)),
		}),
	}
	return a
}

// BuyRAM represents the `eosio.system::buyram` action.
type Deposit struct {
	Player   eos.AccountName `json:"from"`
	Quantity eos.Asset       `json:"quantity"` // specified in EOS
}

func deposit(api *eos.API, player eos.AccountName, amount uint64) {
	actions := []*eos.Action{NewDeposit(player, amount)}
	resp, err := api.SignPushActions(actions...)
	if err == nil {
		data, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println("Ok:", string(data))
	} else {
		fmt.Println("err:", err)
	}
}

func getEos() *eos.API {
	//eos.ch
	api := eos.New("http://jungle.cryptolions.io:38888")
	key := "5J1qniFourzMsbTz1YVuYPRoeGvvivBdsA7yuGwk2sGECWmWZVS"
	signer := eos.NewKeyBag()
	signer.ImportPrivateKey(key)
	//api.c
	api.SetSigner(signer)
	return api
}

func newReveal(player eos.AccountName, rolls uint32) *eos.Action {
	a := &eos.Action{
		Account: player,
		Name:    "reveal",
		Authorization: []eos.PermissionLevel{
			{
				Actor:      kConstractAccount,
				Permission: ("active"),
			},
		},
		ActionData: eos.NewActionData(Reveal{
			From: player,
			Bet:  rolls,
		}),
	}
	return a
}

// BuyRAM represents the `eosio.system::buyram` action.
type Reveal struct {
	From eos.AccountName `json:"from"`
	Bet  uint32          `json:"bet"` // specified in EOS
}

func reveal(api *eos.API, player eos.AccountName, rolls uint32) error {
	actions := []*eos.Action{newReveal(player, rolls)}
	_, err := api.SignPushActions(actions...)
	if err != nil {
		fmt.Println("reveal failed")
		return err
	}

	resp, e := api.GetTableRows(eos.GetTableRowsRequest{
		Code:       kConstractAccount,
		Scope:      kConstractAccount,
		Table:      "account",
		LowerBound: "1000",
		UpperBound: "-1",
		Limit:      100000,
		JSON:       true,
	})
	if e != nil {
		return e
	}

	type result struct {
		Owner   string `json:"owner"`
		Balance uint32 `json:"balance"`
		Bet     uint32 `json:"bet_ret"`
	}
	r := []result{}
	e2 := resp.JSONToStructs(&r)
	if e2 != nil {
		return e2
	}
	fmt.Println(r)
	return nil
}

func NewPayAndClean(player eos.AccountName) *eos.Action {
	a := &eos.Action{
		Account: player,
		Name:    "payandclean",
		Authorization: []eos.PermissionLevel{
			{Actor: kConstractAccount, Permission: ("active")},
		},
		ActionData: eos.NewActionData(PayAndClean{
			Player: player,
		}),
	}
	return a
}

// BuyRAM represents the `eosio.system::buyram` action.
type PayAndClean struct {
	Player eos.AccountName `json:"from"`
}

func getTransaction(api *eos.API, id string, block_id uint32) (error, string, string, float32) {
	resp := eos.TransactionResp{}
	e := api.Call("history", "get_transaction", eos.M{"id": id, "block_num_hint": block_id}, &resp)
	if e != nil {
		return e, "", "", 0
	}
	actions := resp.Transaction.Transaction.Actions
	if actions == nil || len(actions) < 1 {
		return errors.New("invalid actions"), "", "", 0
	}
	data := actions[0].ActionData.Data
	switch d := data.(type) {
	case map[string]interface{}:
		from, ok := d["from"]
		to, ok2 := d["to"]
		quantity, ok3 := d["quantity"]
		if !ok || !ok2 || !ok3 {
			return errors.New("not transfer transaction"), "", "", 0
		}
		q2 := strings.TrimRight(quantity.(string), " EOS")
		q3, ex := strconv.ParseFloat(q2, 32)
		if ex != nil {
			return ex, "", "", 0
		}
		return nil, from.(string), to.(string), float32(q3)
	default:
		fmt.Println(d)
	}
	//fmt.Println(data)
	return nil, "", "", 0
}

func payAndClean(api *eos.API, player eos.AccountName) {
	actions := []*eos.Action{NewPayAndClean(player)}
	resp, err := api.SignPushActions(actions...)
	if err == nil {
		data, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println("Ok:", string(data))
	} else {
		fmt.Println("err:", err)
	}
}

func Test_deposit(t *testing.T) {
	api := getEos()

	account, e := api.GetAccount("qwerasdfvcxz")
	if e != nil {
		t.Error(e)
	}
	fmt.Println(account.CoreLiquidBalance, account.RAMQuota, account.RAMUsage, account.CPULimit, account.NetLimit)

	deposit(api, "qweasdzxcvfr", 1)
	fmt.Println(reveal(api, "qweasdzxcvfr", 50))
	payAndClean(api, "qweasdzxcvfr")
	//fmt.Println(getTransaction(api, "6898eb63e05d1266f51bfcc7f23b508b08625da05c97568a246779e7a0d0eb62", 21940752))
	t.Error("x")
}

func Test_gettrans(t *testing.T) {
	api := getEos()
	id := "6898eb63e05d1266f51bfcc7f23b508b08625da05c97568a246779e7a0d0eb62"
	out := eos.TransactionResp{}
	err := api.Call("history", "get_transaction", eos.M{"id": id, "block_num": 21940752, "block_num_hint": 21940752}, &out)
	fmt.Println(err, "\n\n\n", out)
	t.Error("x")
}

func Test_map(t *testing.T) {
	m := make(map[string]interface{})
	m["x"] = "y"
	m["z"] = "y"
	fmt.Println(m)
	t.Error(m)
}
