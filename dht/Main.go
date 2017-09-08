package dht

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"time"
)

const PORT = 1234

var seed = rand.NewSource(time.Now().UnixNano() % 1000000)
var randSource = rand.New(seed)

func pollChan(ch chan string) {
	for {
		select {
		case str := <-ch:
			fmt.Println(" -- [CH] Recvd from ch: " + str)
		}
	}
}

func BroadcastUpdateWrapper(server *rpc.Client, node DhtNode) AllNodeInfo {
	var reply AllNodeInfo
	if server == nil {
		fmt.Println("!!! WARN: Called BroadcastUpdateWrapper with nil server")
		return AllNodeInfo{}
	}
	broadcastMsg := BroadcastUpdateMsg{NewNode: node,
		RoutedNodes: make(map[int]bool), MsgId: randSource.Intn(90000)}
	broadcastMsg.RoutedNodes[node.NodeId] = true
	fmt.Println(" -- Calling BroadcastUpdate...")
	err := server.Call("DhtServer.BroadcastUpdate", &broadcastMsg, &reply)
	if err != nil {
		log.Fatal("Error calling BroadcastMessage:", err)
		return AllNodeInfo{}
	}
	return reply
}

func debugGet(server *rpc.Client, nodeId int) {

	fmt.Println("Calling client..")
	query := Query{NodeId: nodeId}
	var reply GetResponseType
	err := server.Call("DhtServer.Get", &query, &reply)
	if err != nil {
		log.Fatal("Error calling Get:", err)
		return
	}
	fmt.Println("Got reply from server: ", reply.String())
}

func DialUpdateWrapper(destinationNode DhtNode, updateNode DhtNode) {
	destRpcServer, err := DialHttpTimeout(
		destinationNode.IpAddr.String()+":"+strconv.Itoa(destinationNode.Port), 2)
	if err != nil {
		fmt.Println(" -- Update dial error")
		return
	}
	UpdateWrapper(destRpcServer, updateNode)
}

func UpdateWrapper(server *rpc.Client, updateNode DhtNode) {
	var replyUpdate int
	c := make(chan error, 1)
	go func() { c <- server.Call("DhtServer.Update", &updateNode, &replyUpdate) }()
	select {
	case err := <-c:
		if err != nil {
			log.Println("Error calling Update:", err)
		}
		return
	case <-time.After(2 * time.Second):
		fmt.Println("* Update timeout *")
		return
	}

}

func GetIpString() (string, bool) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Error calling InterfaceAddrs(): " + err.Error() + "\n")
		return "", false
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Println("Got ip: " + ipnet.IP.String() + "\n")
				return ipnet.IP.String(), true
			} else {
				//fmt.Println(" -- nil")
			}
		}
	}

	return "", false
}

func Contains(arr []int, tgt int) bool {
	for _, elem := range arr {
		if elem == tgt {
			return true
		}
	}
	return false
}

/**
Remove old value in arr, replace with newVal; maintian length under $size
*/
func UpdateValueWithReplacement(arr []int, newVal int, size int) {
	if len(arr) < size || len(arr) < 1 {
		arr = append(arr, newVal)
		return
	}
	prev := arr[0]
	for i := 1; i < len(arr); i++ {
		tmp := arr[i]
		arr[i] = prev
		prev = tmp
	}
	arr[0] = newVal
}
