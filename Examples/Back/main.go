package main

import (
	"Gamble/fs"
	"encoding/json"
	"errors"
	"fmt"
	"gamble/helpers"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	eos "github.com/eoscanada/eos-go"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type GambleRound struct {
	Rolls    uint32
	Money    float32
	User     string
	TxID     string
	BlockNum uint32
}

type ClientManager struct {
	sync.RWMutex
	MailBoxes map[string]chan []byte
	ApiEos    *eos.API
}

func NewClientManager() *ClientManager {
	api := eos.New("http://jungle.cryptolions.io:38888")
	key := "5J1qniFourzMsbTz1YVuYPRoeGvvivBdsA7yuGwk2sGECWmWZVS"
	signer := eos.NewKeyBag()
	signer.ImportPrivateKey(key)
	api.SetSigner(signer)

	return &ClientManager{
		MailBoxes: make(map[string]chan []byte),
		ApiEos:    api,
	}
}

func (p *ClientManager) NewMailBox(key string) chan []byte {
	p.Lock()
	defer p.Unlock()

	c := make(chan []byte)
	p.MailBoxes[key] = c
	return c
}

func (p *ClientManager) RemoveMailBox(key string) {
	p.Lock()
	defer p.Unlock()
	delete(p.MailBoxes, key)
}

func (p *ClientManager) Dispatcher(r []byte) {
	p.RLock()
	defer p.RUnlock()
	fmt.Println(len(p.MailBoxes))
	for _, v := range p.MailBoxes {
		v <- r
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = NewClientManager()

func schedule(money float32, rolls uint32, user string, txid string, blocknum string) error {
	blockN, ex := strconv.Atoi(blocknum)
	if ex != nil {
		return ex
	}

	r := &GambleRound{
		Rolls:    rolls,
		Money:    money,
		User:     user,
		TxID:     txid,
		BlockNum: uint32(blockN),
	}

	go func() {
		bs, e := play2(r)
		if e == nil {
			clients.Dispatcher(bs)
		} else {
			fmt.Println(e)
		}
		//clients.Dispatcher(bs)
	}()

	return nil
}

func IsTransactionHandled(txid string) bool {
	return false
}

func play2(round *GambleRound) ([]byte, error) {
	type JsonResult struct {
		Time       string
		User       string
		RollsUnder int32
		Rolls      int32
		BetMoney   float32
		Payout     float32
		TxID       string
	}
	r := JsonResult{}

	//r.TxID = round.TxID
	r.Time = time.Now().Format("2006/01/02 15:04:05")
	r.User = round.User
	r.RollsUnder = int32(round.Rolls)
	r.BetMoney = round.Money
	if IsTransactionHandled(round.TxID) {
		return nil, errors.New("already processed transaction")
	}
	api := helpers.GetEos()
	e, from, to, money := helpers.EosGetTransaction(api, round.TxID, round.BlockNum)
	if e != nil {
		e, from, to, money = helpers.EosGetTransaction(api, round.TxID, round.BlockNum+1)
		if e != nil {
			return nil, e
		}
	}

	if from != round.User {
		return nil, errors.New("Invalid from user")
	}
	if to != helpers.KConstractAccount {
		return nil, errors.New("invalid to user")
	}
	if money != round.Money {
		return nil, errors.New("invalid money")
	}
	dicetxid, e2 := helpers.EosPlayDice(api, eos.AN(round.User), int64(round.Money*10000), uint64(round.Rolls))
	if e2 != nil {
		if strings.Contains(e2.Error(), "Transaction took too long") {
			//ot try again
			dicetxid, e2 = helpers.EosPlayDice(api, eos.AN(round.User), int64(round.Money*10000), uint64(round.Rolls))
			if e2 != nil {
				return nil, e2
			}
		} else {
			return nil, e2
		}
	}

	r.TxID = dicetxid
	e3, rs := helpers.EosGetRevealResult(api)
	if e3 != nil {
		return nil, e3
	}
	if rs == nil || len(rs) < 1 {
		return nil, errors.New("Get reveal result failed")
	}

	r.Rolls = int32(rs[0].Bet)
	r.Payout = float32(rs[0].Balance) / 10000
	/*if r.Rolls < r.RollsUnder {
		//win
		var accurateX = float32(100.0/(r.RollsUnder-1)) * 0.98
		r.Payout = (accurateX * round.Money)
	} */
	return json.Marshal(r)
}

//eosApi := eos.New("xx");

func play(round *GambleRound) ([]byte, error) {
	type JsonResult struct {
		Time       string
		User       string
		RollsUnder int32
		Rolls      int32
		BetMoney   float32
		Payout     float32
	}
	r := JsonResult{}
	r.Rolls = rand.Int31n(100) + 1
	r.Time = time.Now().Format("2006/01/02 15:04:05")
	r.User = round.User
	r.RollsUnder = int32(round.Rolls)
	r.BetMoney = round.Money
	r.Payout = 0
	if r.Rolls < r.RollsUnder {
		//win
		var accurateX = float32(100.0/(r.RollsUnder-1)) * 0.98
		r.Payout = (accurateX * round.Money)
	}
	return json.Marshal(r)
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

func websocketHandle(w http.ResponseWriter, req *http.Request) {
	c, e := upgrader.Upgrade(w, req, nil)
	if e != nil {
		w.Write([]byte(e.Error()))
		return
	}

	defer c.Close()
	key := req.RemoteAddr
	m := clients.NewMailBox(key)
	defer clients.RemoveMailBox(key)

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	go func() {
		defer func() {
			fmt.Println("read goroutine: end")
		}()
		c.SetReadLimit(maxMessageSize)
		c.SetReadDeadline(time.Now().Add(pongWait))
		c.SetPongHandler(func(string) error {
			c.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
					fmt.Println("error:", err)
				}
				break
			}
		}
	}()

	for {
		select {
		case message := <-m:
			err := c.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				fmt.Println(err)
				break
			}
		case <-ticker.C:
			if err := c.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				fmt.Println(err)
				break
			}
		}
	}
}

func gamble(w http.ResponseWriter, req *http.Request) {
	//req.RequestURI.
	url, e := url.Parse(req.RequestURI)
	if e != nil {
		w.Write([]byte(e.Error()))
		return
	}

	query := url.Query()
	rolls := query.Get("rolls")
	money := query.Get("money")
	username := query.Get("username")
	txid := query.Get("txid")
	blockid := query.Get("blockid")
	fmoney, e2 := strconv.ParseFloat(money, 32)
	irolls, e3 := strconv.ParseUint(rolls, 10, 32)
	if e2 != nil {
		w.Write([]byte(e2.Error()))
		return
	}

	if e3 != nil {
		w.Write([]byte(e3.Error()))
		return
	}
	e4 := schedule(float32(fmoney), uint32(irolls), username, txid, blockid)
	if e4 != nil {
		w.Write([]byte(e4.Error()))
	}
}

func main() {

	http.HandleFunc("/websocket", websocketHandle)
	http.HandleFunc("/gamble", gamble)
	http.Handle("/", http.FileServer(http.Dir("static")))
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func main2() {
	memfs, err := fs.New("static")
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.StaticFS("/static", memfs)

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "static/index.html")
	})
	router.POST("/gamble", func(c *gin.Context) {

	})

	//router.Run(":" + helpers.GetConfigPort())
	router.Run(":80")
}
