package main

import (
	log "github.com/sirupsen/logrus"
	"encoding/json"
	"os"
	"fmt"
	"securecommon"
	"net"
	"time"
	paillier "github.com/roasbeef/go-go-gadget-paillier"
	"crypto/rand"
	"strconv"
	"math/big"
)


type Client struct {
	Address string
	Id      string
	Status  int
	CollChan chan common.CoordValue
	PubKey paillier.PublicKey
	Partner1 common.Host
	Partner2 common.Host
	CoordValue []common.CoordValue
	ExitChan chan int

}

var client Client



func main(){
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage:%s id", os.Args)
		os.Exit(1)
	}

	serverAddressStr := fmt.Sprintf("%s:%s",common.Conf.Controller.Address, common.Conf.Controller.Port)
	serverAddr, err := net.ResolveTCPAddr("tcp", serverAddressStr)
	common.CheckError(err)
	serverCon, err := net.DialTCP("tcp", nil, serverAddr)
	common.CheckError(err)
	go handleServerConn(serverCon)

	client = Client{}
	client.Status = 0
	client.Id = os.Args[1]
	client.CollChan = make(chan common.CoordValue)
	client.CoordValue = make([]common.CoordValue, 0)
	client.ExitChan = make (chan int)

	var org common.Host
	if common.Conf.Org1.ID == client.Id {
		org = common.Conf.Org1
		client.Partner1 = common.Conf.Org2
		client.Partner2 = common.Conf.Org3
	} else if common.Conf.Org2.ID == client.Id {
		org = common.Conf.Org2
		client.Partner1 = common.Conf.Org1
		client.Partner2 = common.Conf.Org3
	} else {
		org = common.Conf.Org3
		client.Partner1 = common.Conf.Org1
		client.Partner2 = common.Conf.Org2
	}

	client.Address = fmt.Sprintf("%s:%s",org.Address, org.Port)

	udpAddr, err := net.ResolveUDPAddr("udp", client.Address)
	common.CheckError(err)
	conn, err := net.ListenUDP("udp", udpAddr)
	common.CheckError(err)

	defer conn.Close()
	fmt.Println("client udp listenring started..")

	go handleListen(conn)

	select {
		case exitCode := <- client.ExitChan:
			fmt.Println("Client read to exit..")
			os.Exit(exitCode)
	}

}

//处理合作伙伴的请求
func handleListen(conn *net.UDPConn){
	for {
		log.Info("Wait Receive data from partner")
		buf := make([]byte, 1024)
		dataLen, udpAddr, err := conn.ReadFromUDP(buf)
		common.CheckError(err)
		log.Infof("Receive coord data from partner %s",udpAddr.String())
		var msg common.Msg
		err = json.Unmarshal(buf[:dataLen], & msg)
		log.Infof("Receive coord data from partner %s, cmd=%d",msg.SourceId,msg.Cmd)
		common.CheckError(err)
		if msg.Cmd != 2 {
			return
		}
		var coordData common.CoordValue
		err = json.Unmarshal([]byte(msg.Data), &coordData)

		client.CollChan <- coordData
	}

}

//处理服务端的请求
func handleServerConn(serverConn *net.TCPConn){
	var msg common.Msg
	msg.SourceId = client.Id
	msg.Cmd = 0
	msg.DestId = "0"
	msg.Time = time.Now()
	msgStr, err := json.Marshal(msg)
	common.CheckError(err)

	serverConn.Write(msgStr)

	buf := make([]byte, 1024)
	var number int32
	for {
		log.Info("Wait command info from server ...")
		dataLen, err := serverConn.Read(buf)
		common.CheckError(err)
		err = json.Unmarshal(buf[:dataLen], &msg)
		common.CheckError(err)
		if msg.Cmd == 1  && msg.DestId == client.Id {
			err = json.Unmarshal([]byte(msg.Data), &client.PubKey)
			common.CheckError(err)
			log.Info("Accept New Public key from server ...")
			//log.Info("Please input data for secure compute:")

			//fmt.Scanf("%d", &number)
			t1, _ := rand.Int(rand.Reader, big.NewInt(128))
			number = int32(t1.Uint64())
			log.Info("+++++++++++++++++++++++++++++++++")
			log.Infof("************Customer Wealth value in Org[%s]: %d(random)   ***********",client.Id, number)
			log.Info("+++++++++++++++++++++++++++++++++")

			go handleDataDispatch(serverConn, number)
			var exitStatus int = 0
			for  {
				select {
				case coordValue := <- client.CollChan:
					log.Info("Received  coord data  from partner ...")
					client.CoordValue = append(client.CoordValue, coordValue)
					if len(client.CoordValue) == 3 {
						log.Info(" Finish receive data from partner,start compute total and send to center server")
						computeTotal(serverConn)
						exitStatus = 1
						break
					}
				}
				if exitStatus == 1 {
					log.Info("")
					log.Info("++End of this round of calculation++")
					log.Info("")
					break
				}
			}
		} else if msg.Cmd == 4 {
			log.Info("Receive exit command from server ...")
			client.ExitChan <- 1
			client.Status = 1
			break
		}
	}


}

func computeTotal(serverConn *net.TCPConn){
	log.Info("Compute total value,wait..")
	var resultCoord common.CoordValue
	var resultInt int
	resultInt = 0
	for i:= 0; i < len(client.CoordValue);i++{
		resultCoord.Xvalue = client.CoordValue[i].Xvalue
		value, err:= strconv.Atoi(client.CoordValue[i].Yvalue)
		common.CheckError(err)
		resultInt = resultInt + value
	}
	resultCoord.Yvalue = strconv.Itoa(resultInt)
	var msg common.Msg
	msg.Cmd = 3
	msg.DestId = "0"
	msg.SourceId = client.Id
	msg.Time = time.Now()
	coordStr,err := json.Marshal(resultCoord)
	common.CheckError(err)
	msg.Data = string(coordStr)

	msgStr,err := json.Marshal(msg)
	common.CheckError(err)
	log.Info("send total value to controller")
	serverConn.Write(msgStr)

}

func generatePoly(number int32) (x1, y1, x2, y2, x3, y3 int32){
	t1, _ := rand.Int(rand.Reader, big.NewInt(128))
	t2, _ := rand.Int(rand.Reader, big.NewInt(128))
	a := int32(t1.Uint64())
	b := int32(t2.Uint64())

	x1 = 1
	x2 = 2
	x3 = 3

	y1 = a * x1 * x1 + b * x1 + number
	y2 = a * x2 * x2 + b * x2 + number
	y3 = a * x3 * x3 + b * x3 + number

	return x1, y1, x2, y2, x3, y3
}

func sendValue(host common.Host, value common.CoordValue){
	partAddress := fmt.Sprintf("%s:%s",host.Address, host.Port)

	p1Addr, err := net.ResolveUDPAddr("udp", partAddress)
	common.CheckError(err)
	p1Conn,err := net.DialUDP("udp",nil,p1Addr)
	common.CheckError(err)
	var msg common.Msg
	coordValue,err := json.Marshal(value)
	msg.Data = string(coordValue)
	msg.Time = time.Now()
	msg.SourceId = client.Id
	msg.DestId = host.ID
	msg.Cmd = 2
	data, err := json.Marshal(msg)
	common.CheckError(err)
	p1Conn.Write(data)
}

func handleDataDispatch(serverConn *net.TCPConn, number int32){
	log.Info("Start Security mulitply Compute...,Generate random param")
	client.CoordValue = append([]common.CoordValue{})
	x1, y1, x2, y2, x3, y3 := generatePoly(number)
	cv1 := common.CoordValue{}
	cv2 := common.CoordValue{}
	cv3 := common.CoordValue{}

	if client.Id == "1" {
		cv1.Xvalue = fmt.Sprintf("%d", x1)
		cv1.Yvalue = fmt.Sprintf("%d", y1)

		cv2.Xvalue = fmt.Sprintf("%d", x2)
		cv2.Yvalue = fmt.Sprintf("%d", y2)

		cv3.Xvalue = fmt.Sprintf("%d", x3)
		cv3.Yvalue = fmt.Sprintf("%d", y3)

	}else if client.Id == "2" {
		cv1.Xvalue = fmt.Sprintf("%d", x2)
		cv1.Yvalue = fmt.Sprintf("%d", y2)

		cv2.Xvalue = fmt.Sprintf("%d", x1)
		cv2.Yvalue = fmt.Sprintf("%d", y1)

		cv3.Xvalue = fmt.Sprintf("%d", x3)
		cv3.Yvalue = fmt.Sprintf("%d", y3)
	}else {
		cv1.Xvalue = fmt.Sprintf("%d", x3)
		cv1.Yvalue = fmt.Sprintf("%d", y3)

		cv2.Xvalue = fmt.Sprintf("%d", x1)
		cv2.Yvalue = fmt.Sprintf("%d", y1)

		cv3.Xvalue = fmt.Sprintf("%d", x2)
		cv3.Yvalue = fmt.Sprintf("%d", y2)
	}
	client.CoordValue = append(client.CoordValue, cv1)
	log.Infof("send to parter :%s:%s",client.Partner1.Address,client.Partner1.Port)
	sendValue(client.Partner1,cv2)
	log.Infof("send to parter :%s:%s",client.Partner2.Address,client.Partner2.Port)
	sendValue(client.Partner2,cv3)
}