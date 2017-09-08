package dht

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"time"
)

const BOOTSTRAP_URL = "http://blobtag.herokuapp.com"
const RECENT_BROADCAST_IDS_LENGTH = 2

var mostRecentBroadcastId = -1

type Query struct {
	NodeId int
}

type GetResponseType struct {
	Node    DhtNode
	Success bool
}

func (resp GetResponseType) String() string {
	return "Found? " + strconv.FormatBool(resp.Success) +
		". Node: " + resp.Node.String()
}

type AllNodeInfo struct {
	//node IDs that this message has already been routed to
	NodeMap map[int]DhtNode
}

type BroadcastUpdateMsg struct {
	NewNode DhtNode
	MsgId   int
	//node IDs that this message has already been routed to
	RoutedNodes map[int]bool
}

type DhtServer struct {
	isInitialized bool

	//update messages are sent on this channel
	UpdateChan chan string
	//Node corresponding to this server
	SelfNode       DhtNode
	SelfConnection *rpc.Client
	//map node ID -> RPC client object
	//maps nodeId to node
	NodeMap map[int]DhtNode
	ZoneMap map[int]DhtNode
}

//Initialize nodeMap; Launch RPC server
func (ds *DhtServer) Init(selfNode DhtNode) error {
	ds.NodeMap = make(map[int]DhtNode)
	ds.ZoneMap = make(map[int]DhtNode)
	ds.SelfNode = selfNode
	ds.NodeMap[selfNode.NodeId] = selfNode
	ds.ZoneMap[selfNode.NodeId] = selfNode
	rpc.Register(ds)
	gob.Register(Query{})
	gob.Register(GetResponseType{})
	gob.Register(AllNodeInfo{})
	rpc.HandleHTTP()

	fmt.Printf("Calling Listen on port %v\n", selfNode.Port)
	l, e := net.Listen("tcp", ":"+strconv.Itoa(selfNode.Port))
	if e != nil {
		log.Fatal("listen error:", e)
		return e
	}

	go http.Serve(l, nil)
	var err error
	ds.SelfConnection, err = rpc.DialHTTP("tcp", selfNode.IpAddr.String()+":"+strconv.Itoa(selfNode.Port))

	if err != nil {
		log.Fatal("Dialing self connection: ", err)
		return err
	}
	fmt.Printf(" -- Initialized selfConnection.\n")

	httpResponse, err := http.Get(BOOTSTRAP_URL + "/update/upd?nodeId=" +
		strconv.Itoa(selfNode.NodeId) + "&addr=" +
		selfNode.IpAddr.String() + ":" + strconv.Itoa(selfNode.Port))

	if err != nil {
		fmt.Println("!! Error in get request to update bootstrap node: " + err.Error())
		return err
	}
	responseJsonStr, err := ioutil.ReadAll(httpResponse.Body)
	// fmt.Println(" -- Bootstrap response string: " + string(responseJsonStr))
	if err != nil {
		fmt.Println("!!! Error reading from httpResponse body")
		return err
	}
	httpResponse.Body.Close()
	bootstrapNodeInfo := NodeInfoHistory{}
	json.Unmarshal(responseJsonStr, &bootstrapNodeInfo)
	ds.isInitialized = true

	//connect to all nodes returned by bootstrap server using selfConnection
	fmt.Printf(" -- bootstrapNodeInfo.NodeMap: %v\n", bootstrapNodeInfo.NodeMap)
	for _, currNode := range bootstrapNodeInfo.NodeMap {
		fmt.Printf(" -- currNode Id: %v; selfNode Id: %v\n", currNode.NodeId, selfNode.NodeId)
		if currNode.NodeId == selfNode.NodeId {
			continue
		}
		makeBootstrapConnection(ds.SelfConnection, currNode)
	}

	allNodes := BroadcastUpdateWrapper(ds.SelfConnection, selfNode)
	fmt.Println(" -- Connected to nodes: ")
	for _, currNode := range allNodes.NodeMap {
		fmt.Printf("\t - %v\n", currNode.String())
	}
	return nil
}

func makeBootstrapConnection(selfConn *rpc.Client, newNode *NodeInfo) bool {

	server, err := DialHttpTimeout(newNode.NetAddr, 2) //2 secs

	if err != nil {
		fmt.Printf("makeBootstrapConnection: Could not dial %v: %v",
			newNode.NetAddr, err.Error())
		return false
	}
	var neighborNode DhtNode
	var empty int = 0
	//timed RPC call
	c := make(chan error, 1)
	go func() { c <- server.Call("DhtServer.GetNodeInfo", &empty, &neighborNode) }()
	select {
	case err := <-c:
		err = server.Call("DhtServer.GetNodeInfo", &empty, &neighborNode)
		server.Close()
		if err != nil {
			fmt.Println("makeBootstrapConnection: Could not rpcCall %v: %v",
				newNode.NetAddr, err.Error())
			return false
		}
		UpdateWrapper(selfConn, neighborNode)
		return true
	case <-time.After(2 * time.Second):
		fmt.Println(" -- timeout:" + neighborNode.IpAddr.String())
		return false
	}

	return true
}

/**
Search this node for query.NodeId. If not found, forward to the query to other
nodes
*/
func (ds DhtServer) Get(query *Query, reply *GetResponseType) error {
	storedNode, isPresent := ds.NodeMap[query.NodeId]

	if isPresent {
		reply.Node = storedNode
		reply.Success = true
	} else {
		reply.Success = false
	}

	return nil
}

func (ds DhtServer) Update(updateNode *DhtNode, reply *int) error {

	fmt.Printf(" -- Received update at %v, %v about %v, %v\n", ds.SelfNode.NodeId, ds.SelfNode.BlobInfo.Zone,
		updateNode.NodeId, updateNode.BlobInfo.Zone)
	if updateNode.BlobInfo.Zone == ds.SelfNode.BlobInfo.Zone {
		ds.ZoneMap[updateNode.NodeId] = *updateNode
		jsonNodeRepr, err := json.Marshal(updateNode.BlobInfo)
		if err != nil {
			fmt.Println("!!! Error Marshalling blobinfo in BroadcastUpdate: " + err.Error())
		}
		fmt.Println(" --  --Sending update to UI!")
		ds.UpdateChan <- "[" + string(jsonNodeRepr) + "]"
	} else {
		delete(ds.ZoneMap, updateNode.NodeId)
	}

	ds.NodeMap[updateNode.NodeId] = *updateNode
	if updateNode.NodeId == ds.SelfNode.NodeId {
		//update selfnode
		ds.SelfNode = *updateNode
	}
	fmt.Println(" -- Update successful!")
	*reply = 0
	return nil
}

func (ds DhtServer) GetNodeInfo(empty *int, reply *DhtNode) error {
	*reply = ds.SelfNode
	return nil
}

/**
TODO change to BroadcastUpdate
Called by a new node to join the network
*/
func (ds DhtServer) BroadcastUpdate(
	bmsg *BroadcastUpdateMsg, reply *AllNodeInfo) error {

	fmt.Printf(" -- Brdcst\n")
	// bmsg.MsgId, mostRecentBroadcastId == bmsg.MsgId)
	if mostRecentBroadcastId == bmsg.MsgId {
		//already received this message; ignore
		*reply = AllNodeInfo{NodeMap: ds.NodeMap}
		fmt.Printf("\t.\n")
		return nil
	}
	fmt.Printf(" -- bcast: selfZone: %v, %v; updatedNodeZone: %v, %v\n",
		ds.SelfNode.NodeId, ds.SelfNode.BlobInfo.Zone, bmsg.NewNode.NodeId,
		bmsg.NewNode.BlobInfo.Zone)
	mostRecentBroadcastId = bmsg.MsgId
	//store message ID as 'seen'
	if bmsg.NewNode.BlobInfo.Zone == ds.SelfNode.BlobInfo.Zone {
		/* Forward new information to front end if in same zone*/
		jsonNodeRepr, err := json.Marshal(bmsg.NewNode.BlobInfo)
		if err != nil {
			fmt.Println("!!! Error Marshalling blobinfo in BroadcastUpdate: " + err.Error())
		}
		//JSON to front end should be in list form
		//TODO do this only if not selfNode
		ds.UpdateChan <- "[" + string(jsonNodeRepr) + "]"
		fmt.Printf(" -- Updated UI. selfZone: %v; updateNodeZone: %v\n",
			ds.SelfNode.BlobInfo.Zone, bmsg.NewNode.BlobInfo.Zone)
		if bmsg.NewNode.NodeId != ds.SelfNode.NodeId {

			fmt.Println(" -- Responding with update..")
			ds.SelfNode.UpdateBlobPositionEstimate()
			if bmsg.NewNode.BlobInfo.ZoneChanged {
				//only update if zone changed
				DialUpdateWrapper(bmsg.NewNode, ds.SelfNode)
			}
		}
	}

	/* Forward to all other nodes in network */
	broadcastCount := 0
	//create new BroadcastMessage to be forwarded
	newMessage := BroadcastUpdateMsg{NewNode: bmsg.NewNode,
		RoutedNodes: bmsg.RoutedNodes, MsgId: bmsg.MsgId}
	newMessage.RoutedNodes[ds.SelfNode.NodeId] = true
	alreadyRoutedIds := make([]int, 0)
	routedToIds := make([]int, 0)
	for routedNodeId, _ := range bmsg.RoutedNodes {
		newMessage.RoutedNodes[routedNodeId] = true
		alreadyRoutedIds = append(alreadyRoutedIds, routedNodeId)
	}

	for _, node := range ds.NodeMap {
		_, hasBeenRouted := bmsg.RoutedNodes[node.NodeId]
		if node.IpAddr.Equal(ds.SelfNode.IpAddr) &&
			node.Port == ds.SelfNode.Port {
			continue
		} else if hasBeenRouted {
			continue
		}
		broadcastCount++
		routedToIds = append(routedToIds, node.NodeId)
		returnedMap := sendBroadcastUpdate(node, newMessage)
		updatedNode, isPresent := returnedMap.NodeMap[node.NodeId]
		if !isPresent {
			fmt.Println("!!! WARNING: BroadcastUpdate nodemap id not present")
			continue
		}
		ds.NodeMap[node.NodeId] = updatedNode
		if ds.SelfNode.BlobInfo.Zone == (updatedNode.BlobInfo.Zone) {
			ds.ZoneMap[node.NodeId] = updatedNode
		} else {
			delete(ds.ZoneMap, updatedNode.NodeId)
		}
	}

	ds.NodeMap[bmsg.NewNode.NodeId] = bmsg.NewNode
	if bmsg.NewNode.NodeId == ds.SelfNode.NodeId {
		//update selfnode
		ds.SelfNode = bmsg.NewNode
	}
	*reply = AllNodeInfo{NodeMap: ds.NodeMap}
	return nil
}

/**
Uupdates all nodes in my zone
*/
func (ds DhtServer) ZoneUpdate(
	bmsg *BroadcastUpdateMsg, reply *AllNodeInfo) error {

	/* Forward to all other nodes in network */
	broadcastCount := 0
	//create new BroadcastMessage to be forwarded
	newMessage := BroadcastUpdateMsg{NewNode: bmsg.NewNode,
		RoutedNodes: bmsg.RoutedNodes}
	newMessage.RoutedNodes[ds.SelfNode.NodeId] = true

	for _, node := range ds.ZoneMap {
		if node.IpAddr.Equal(ds.SelfNode.IpAddr) &&
			node.Port == ds.SelfNode.Port {
			continue
		} else if _, hasBeenRouted := bmsg.RoutedNodes[node.NodeId]; hasBeenRouted {
			continue
		}
		broadcastCount++
		sendZoneUpdate(node, newMessage)
	}

	ds.NodeMap[bmsg.NewNode.NodeId] = bmsg.NewNode
	if bmsg.NewNode.BlobInfo.Zone == (ds.SelfNode.BlobInfo.Zone) {
		ds.ZoneMap[bmsg.NewNode.NodeId] = bmsg.NewNode
		/* Forward new information to front end */
		jsonNodeRepr, err := json.Marshal(bmsg.NewNode.BlobInfo)
		if err != nil {
			fmt.Println("!!! Error Marshalling blobinfo in BroadcastUpdate: " + err.Error())
		}
		// JSON to front end should be in list form
		ds.UpdateChan <- "[" + string(jsonNodeRepr) + "]"
	} else {
		delete(ds.ZoneMap, bmsg.NewNode.NodeId)
	}
	*reply = AllNodeInfo{NodeMap: ds.ZoneMap}
	fmt.Printf(" -- ZoneBroadcasted to %d nodes.\n", broadcastCount)
	return nil
}

func sendZoneUpdate(node DhtNode, bmsg BroadcastUpdateMsg) AllNodeInfo {
	fmt.Println(" -- Dialing client")
	server, err := rpc.DialHTTP("tcp",
		node.IpAddr.String()+":"+strconv.Itoa(node.Port))

	if err != nil {
		fmt.Println("error dialing:", err.Error())
	}

	var response AllNodeInfo
	fmt.Println("Calling client..")
	err = server.Call("DhtServer.ZoneUpdate", &bmsg, &response)
	server.Close()
	if err != nil {
		fmt.Println("Error calling ZoneUpdate:", err.Error())
		return AllNodeInfo{}
	}
	return response
}

func sendBroadcastUpdate(node DhtNode, bmsg BroadcastUpdateMsg) AllNodeInfo {
	server, err := rpc.DialHTTP("tcp",
		node.IpAddr.String()+":"+strconv.Itoa(node.Port))

	if err != nil {
		fmt.Println(" error dialing:", err)
		return AllNodeInfo{}
	}

	var response AllNodeInfo
	err = server.Call("DhtServer.BroadcastUpdate", &bmsg, &response)
	server.Close()
	if err != nil {
		return AllNodeInfo{}
	}
	return response
}

type DialHttpResult struct {
	Client *rpc.Client
	Err    error
}

func DialHttpTimeout(addr string, timeoutSecs time.Duration) (*rpc.Client, error) {

	c := make(chan DialHttpResult, 1)
	go func() {
		server, err := rpc.DialHTTP("tcp", addr)
		c <- DialHttpResult{Client: server, Err: err}
	}()

	select {
	case result := <-c:
		return result.Client, result.Err
	case <-time.After(timeoutSecs * time.Second):
		return nil, errors.New("timeout")
	}
}
