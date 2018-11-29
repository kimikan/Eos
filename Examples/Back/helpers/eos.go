package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	eos "github.com/eoscanada/eos-go"
)

const (
	KConstractAccount = "qwerasdfvcxz"
)

func NewDeposit(player eos.AccountName, eosQuantity uint64) *eos.Action {
	a := &eos.Action{
		Account: KConstractAccount,
		Name:    "deposit",
		Authorization: []eos.PermissionLevel{
			{Actor: KConstractAccount, Permission: eos.PN("active")},
		},
		ActionData: eos.NewActionData(Deposit{
			From:     player,
			Quantity: uint32(eosQuantity),
		}),
	}
	return a
}

// BuyRAM represents the `eosio.system::buyram` action.
type Deposit struct {
	From     eos.AccountName `json:"from"`
	Quantity uint32          `json:"quantity"` // specified in EOS
}

func EosDeposit(api *eos.API, player eos.AccountName, amount uint64) error {
	actions := []*eos.Action{NewDeposit(player, amount)}
	resp, err := api.SignPushActions(actions...)
	if err == nil {
		data, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println("Ok:", string(data))
		return nil
	}
	return err
}

func NewEos() *eos.API {
	api := eos.New("http://jungle.cryptolions.io:38888")
	key := "5J1qniFourzMsbTz1YVuYPRoeGvvivBdsA7yuGwk2sGECWmWZVS"
	signer := eos.NewKeyBag()
	signer.ImportPrivateKey(key)
	//api.c
	api.SetSigner(signer)
	return api
}

var gEos *eos.API

func GetEos() *eos.API {
	if gEos == nil {
		gEos = NewEos()
	}

	return gEos
}

type RevealResult struct {
	Balance uint32 `json:"balance"`
	Bet     uint32 `json:"bet_ret"`
}

func EosGetRevealResult(api *eos.API) (error, []RevealResult) {
	resp, e := api.GetTableRows(eos.GetTableRowsRequest{
		Code:  KConstractAccount,
		Scope: KConstractAccount,
		Table: "account",
		JSON:  true,
	})
	if e != nil {
		return e, nil
	}

	r := []RevealResult{}
	e2 := resp.JSONToStructs(&r)
	if e2 != nil {
		return e2, nil
	}
	//fmt.Println(r)
	return nil, r
}

func newReveal(player eos.AccountName, rolls uint32) *eos.Action {
	a := &eos.Action{
		Account: KConstractAccount,
		Name:    "reveal",
		Authorization: []eos.PermissionLevel{
			{
				Actor:      KConstractAccount,
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

func EosReveal(api *eos.API, player eos.AccountName, rolls uint32) error {
	actions := []*eos.Action{newReveal(player, rolls)}
	resp, err := api.SignPushActions(actions...)
	if err != nil {
		fmt.Println("reveal failed")
		return err
	}
	fmt.Println(resp)
	return nil
}

func NewPayAndClean(player eos.AccountName) *eos.Action {
	a := &eos.Action{
		Account: KConstractAccount,
		Name:    "payandclean",
		Authorization: []eos.PermissionLevel{
			{Actor: KConstractAccount, Permission: ("active")},
		},
		ActionData: eos.NewActionData(PayAndClean{
			Player: player,
		}),
	}
	return a
}

// BuyRAM represents the `eosio.system::buyram` action.
type PayAndClean struct {
	Player eos.AccountName `json:"player"`
}

func EosGetTransaction(api *eos.API, id string, block_id uint32) (error, string, string, float32) {
	resp := eos.TransactionResp{}
	e := api.Call("history", "get_transaction", eos.M{
		"id":             id,
		"block_num_hint": block_id}, &resp)
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

//EosPayAndClean......Api
func EosPayAndClean(api *eos.API, player eos.AccountName) error {
	actions := []*eos.Action{NewPayAndClean(player)}
	resp, err := api.SignPushActions(actions...)
	if err == nil {
		data, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println("Ok:", string(data))
	} else {
		fmt.Println("Err:", err)
	}
	return err
}

func newDice(player eos.AccountName, amount int64, rolls uint64) *eos.Action {
	a := &eos.Action{
		Account: KConstractAccount,
		Name:    eos.ActN("deposit"),
		Authorization: []eos.PermissionLevel{
			{
				Actor:      eos.AN(KConstractAccount),
				Permission: eos.PN("active"),
			},
		},
		ActionData: eos.NewActionData(Dice{
			Player: player,
			Amount: eos.NewEOSAsset(amount),
			//Amount: amount,
			Rolls: rolls,
		}),
	}

	resp, _ := json.Marshal(a.ActionData)
	fmt.Println(string(resp))
	return a
}

//eos.ActionData
// BuyRAM represents the `eosio.system::buyram` action.
type Dice struct {
	Player eos.AccountName `json:"from"`
	Amount eos.Asset       `json:"quantity"` // specified in EOS
	//Amount int64
	Rolls uint64 `json:"bet"` // specified in EOS
}

func EosPlayDice(api *eos.API, player eos.AccountName, amount int64, rolls uint64) (string, error) {
	actions := []*eos.Action{newDice(player, amount, rolls)}
	/*
		chain := eos.SHA256Bytes("038f4b0fc8ff18a4f0842a8f0564611f6e96e8535901dd45e43ac8691a1c4dca")
		opts := eos.TxOptions{}
		opts.FillFromChain(api)
		opts.ChainID = chain
		transaction := eos.NewTransaction(actions, &opts)
		resp, err := api.SignPushTransaction(transaction, opts.ChainID, opts.Compress)
	*/
	resp, err := api.SignPushActions(actions...)
	if err != nil {
		fmt.Println("Play dice failed!")
		return "", err
	}

	fmt.Println(resp)
	return resp.TransactionID, nil
}
