package main

import (
	log "github.com/sirupsen/logrus"
	"encoding/json"
	"fmt"
	"net"
	"time"
	"securecommon"
	paillier "github.com/roasbeef/go-go-gadget-paillier"
	"crypto/rand"
	"strconv"
	"bufio"
	"os"
)

type Server struct {
	Address string
	Conns [] common.Connection
	ClientCount int
	ClientIdChan chan int
}

var server Server



func main(){

	server = Server{}
	server.ClientCount = 0
	server.Address = fmt.Sprintf("%s:%s",common.Conf.Controller.Address, common.Conf.Controller.Port)
	server.Conns = make([]common.Connection, 0)
	server.ClientIdChan = make(chan int)

	tcpAddr, err := net.ResolveTCPAddr("tcp4", server.Address)
	common.CheckError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	common.CheckError(err)
	defer listener.Close()

	go handleListen(listener)
	var option int
	var hasReady int = 0
	for {
		if len(server.Conns) < 3 {
			fmt.Println("Wait client to join,please wait...")
			for  {
				select {
					case clientId := <- server.ClientIdChan:
						fmt.Printf("Client %d join cluster\n", clientId)
						if len(server.Conns) == 3 {
							fmt.Println("All members has joined...")
							hasReady = 1
							break
						}
				}
				if hasReady == 1 {
					break
				}
			}

		}

		fmt.Println("Please input Option number to Continue:")
		fmt.Println("1: Start Collect Wealth Value")
		fmt.Println("2: Quit")

		collectOps := common.CollectOps{}
		collectOps.CollChan = make(chan common.CoordValue)
		collectOps.PrivKey,err = paillier.GenerateKey(rand.Reader, 128)
		common.CheckError(err)

		//option = 1
		//fmt.Scanf("%d", &option)
		stdin := bufio.NewReader(os.Stdin)
		_, _ = fmt.Fscan(stdin, &option)
		stdin.ReadString('\n')
		if option == 2 {
			sendExitMsg()
			break
		} else if option == 1 {
			if len(server.Conns) != 3 {
				fmt.Println("There are currently clients not connected，Please Wait")
				continue
			}
			Collect(collectOps)
		}

	}
	//关闭连接
	for i := 0; i < len(server.Conns) ;i++ {
		server.Conns[i].Conn.Close()
	}


}

func sendExitMsg(){
	var msg common.Msg
	msg.Cmd = 4
	msg.SourceId = "0"
	msg.Time = time.Now()

	for i := 0; i < len(server.Conns); i++ {
		msg.DestId = fmt.Sprintf("%s", server.Conns[i].Id)
		msgByte, err := json.Marshal(msg)
		common.CheckError(err)
		server.Conns[i].Conn.Write(msgByte)
		server.Conns[i].Status = 4
	}
}

func handleListen(listener net.Listener){
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn){
	//defer conn.Close()
	clientConnect := common.Connection{}
	clientConnect.Conn = conn
	buf := make([]byte, 1024)
	dataLen, err := conn.Read(buf)
	common.CheckError(err)
	var msg common.Msg
	err = json.Unmarshal(buf[:dataLen], &msg)
	common.CheckError(err)
	clientConnect.Id ,err = strconv.Atoi(msg.SourceId)
	common.CheckError(err)
	clientConnect.Status = 1
	server.Conns = append(server.Conns, clientConnect)
	log.Infof("Participants %d has joined", clientConnect.Id)
	server.ClientIdChan <- clientConnect.Id

}

func Collect(collectOps common.CollectOps){

	log.Info("start secure compute..")
	var msg common.Msg
	msg.Cmd = 1
	msg.SourceId = "0"
	msg.Time = time.Now()

	log.Info("generate paillier key..")
	pubkey, err := json.Marshal(collectOps.PrivKey.PublicKey)
	common.CheckError(err)
	msg.Data = string(pubkey)
	for i := 0; i < len(server.Conns); i++ {
		msg.DestId = fmt.Sprintf("%d", server.Conns[i].Id)
		msgByte, err := json.Marshal(msg)
		common.CheckError(err)
		server.Conns[i].Conn.Write(msgByte)
		server.Conns[i].Status = 2
		go handleCollect(server.Conns[i], collectOps)
	}
	var coordValues []common.CoordValue
	var exitStatus int = 0
	for  {
		select {
			case coordValue := <- collectOps.CollChan:
				coordValues = append(coordValues,coordValue)
				if len(coordValues) == 3 {
					computeWealthTotal(coordValues, collectOps)
					exitStatus = 1
					break
				}

		}
		if exitStatus == 1 {
			break
		}
	}

}

func handleCollect(connection common.Connection, collectOps common.CollectOps)  {
	buf := make([]byte, 1024)
	dataLen,err := connection.Conn.Read(buf)
	common.CheckError(err)
	var msg common.Msg
	err = json.Unmarshal(buf[:dataLen], &msg)
	common.CheckError(err)
	var coordValue common.CoordValue
	err = json.Unmarshal([]byte(msg.Data), &coordValue)
	common.CheckError(err)
	collectOps.CollChan <- coordValue
	connection.Status = 3

}

func convertCoord(coordValue common.CoordValue)(int,int){
	x,err := strconv.Atoi(coordValue.Xvalue)
	common.CheckError(err)
	y,err := strconv.Atoi(coordValue.Yvalue)
	common.CheckError(err)
	return x,y
}

func computeWealthTotal(coordValues []common.CoordValue, collectOps common.CollectOps){
	//todo 根据三个坐标点，计算常量值
	x1,y1 := convertCoord(coordValues[0])
	x2,y2 := convertCoord(coordValues[1])
	x3,y3 := convertCoord(coordValues[2])

	result := ((y1*x2*x2-y2*x1*x1)*(x2*x3*x3-x3*x2*x2)-(y2*x3*x3-y3*x2*x2)*(x1*x2*x2-x2*x1*x1))/((x2*x2-x1*x1)*(x2*x3*x3-x3*x2*x2)-(x3*x3-x2*x2)*(x1*x2*x2-x2*x1*x1))
	log.Infof("********** Custom Total Wealth :%d **************", result)

}