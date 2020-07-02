package common


import (
	"time"
	"net"
	. "github.com/roasbeef/go-go-gadget-paillier"
	"math/big"
)

type Msg struct {
	Cmd        int `json:"cmd"`        //0:connect,1:start ,2: send to partis,3:send to controller,4:exit
	SourceId   string `json:"sourceid"`
	DestId     string `json:"destid"`
	Data       string `json:"data"`
	Time       time.Time `json:"time"`
}

type CoordValue struct {
	Xvalue   string `json:"xvalue"`
	Yvalue   string `json:"yvalue"`
}



type Connection struct {
	Conn    net.Conn
	Id      int
	Status  int    //0:not connect,1:connect,2:发送,3:接收,4:disconnect
}


type CollectOps struct {
	PrivKey *PrivateKey
	WealthValue big.Int
	CollChan chan CoordValue
}
