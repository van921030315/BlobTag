package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"time"

	"github.com/blobtag/dht"
	"github.com/blobtag/webapp"
	"github.com/pkg/browser"
)

const RPC_PORT_IDX = 1

const NODE_ID_IDX = 2
const BOARD_WIDTH = 700
const BOARD_HEIGHT = 500

var seed = rand.NewSource(time.Now().UnixNano() % 1000000)
var randSource = rand.New(seed)

type NetworkInfo struct {
	//list of node addresses in the form <ip>:<port>
	nodeAddrMap map[string]*rpc.Client
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println(" Wrong # of args. Should be ./coordinator <port> <id>")
		return
	}
	rpcPort, err := strconv.Atoi(os.Args[RPC_PORT_IDX])

	if err != nil {
		fmt.Println("!! please enter a valid port!")
		return
	}

	playerName := os.Args[RPC_PORT_IDX+1]

	selfNodeId := randSource.Intn(40000)

	if err != nil {
		fmt.Println("!! please enter a valid id!")
		return
	}

	updateChan := make(chan string, 1000)

	ipStr, success := dht.GetIpString()
	if !success {
		fmt.Println("Could not get IP string, aborting.")
		return
	}
	ipAddr := net.ParseIP(ipStr)

	redComp, blueComp, greenComp := randSource.Intn(256),
		randSource.Intn(256), randSource.Intn(256)
	colorVal := (redComp << 16) | (blueComp << 8) | greenComp
	myBlob := dht.BlobInfo{PlayerName: playerName + strconv.Itoa(selfNodeId), IsAlive: true,
		Color: colorVal, Xsize: 50.0, Ysize: 50.0,
		Xvel: 0.1, Yvel: 0.1, Xpos: float64(randSource.Intn(BOARD_WIDTH)),
		Ypos:        float64(randSource.Intn(BOARD_HEIGHT)),
		Zone:        randSource.Intn(4),
		ZoneChanged: false}

	/* Initialize web server */
	fromClientChan, toClientChan := make(chan string, 100), make(chan string, 100)
	webServer := webapp.WebServer{
		SelfBlobInfo:   myBlob,
		SelfNodeId:     selfNodeId,
		WebServerPort:  strconv.Itoa(rpcPort + 1),
		WebsocketPort:  strconv.Itoa(rpcPort + 2),
		FromClientChan: fromClientChan,
		ToClientChan:   toClientChan}

	webServer.InitServer(strconv.Itoa(rpcPort + 1))

	/* initialize RPC server */
	selfNode := dht.DhtNode{NodeId: selfNodeId, IpAddr: ipAddr, Port: rpcPort, BlobInfo: &myBlob,
		UpdateTimeNanos: time.Now().Unix()}
	dhtServer := dht.DhtServer{UpdateChan: updateChan, NodeMap: make(map[int]dht.DhtNode)}
	dhtServer.Init(selfNode)

	go pollUpdateChan(&webServer, updateChan)
	// go handleTicker(&webServer)
	netInfo := NetworkInfo{nodeAddrMap: make(map[string]*rpc.Client)}
	fmt.Println(" -- Opening browser...")
	browser.OpenURL("http://localhost:" + strconv.Itoa(rpcPort+1) + "/go")
	fmt.Println(" -- done.")
	pollWebClient(&webServer, &netInfo, &dhtServer, &selfNode)

}

//Poll for updates from game
func pollWebClient(webserver *webapp.WebServer, netInfo *NetworkInfo,
	dhtServer *dht.DhtServer, selfNode *dht.DhtNode) {
	for {
		select {
		case blobUpdateJson := <-webserver.FromClientChan:
			updatedBlobInfo := dht.BlobInfo{}
			err := json.Unmarshal([]byte(blobUpdateJson), &updatedBlobInfo)

			selfNode.BlobInfo = &updatedBlobInfo
			fmt.Printf(" -- UI update; zone: %v\n", selfNode.BlobInfo.Zone)
			selfNode.UpdateTimeNanos = time.Now().UnixNano()
			dhtServer.SelfNode.BlobInfo = &updatedBlobInfo
			dhtServer.SelfNode.UpdateTimeNanos = time.Now().UnixNano()
			if err != nil {
				fmt.Println("!!! Error Unmarshalling json: " + err.Error())
				continue
			}
			if dhtServer.SelfConnection != nil {
				dht.BroadcastUpdateWrapper(dhtServer.SelfConnection, *selfNode)
			}
		}
	}
}

//Poll for external updates
func pollUpdateChan(wa *webapp.WebServer, pollChan chan string) {
	for {
		select {
		case val := <-pollChan:
			wa.ToClientChan <- val

		}
	}
}
