package CAN

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"testing"
)

var (
	blob          = NewBlobInfo("Bootstrap", 0, 0, 0)
	ip, _, _      = net.ParseCIDR("127.0.0.1/32")
	bszone        = NewZone(Coordinates{0, 0}, Coordinates{1, 1})
	bootstrapNode = NewDhtNode(Coordinates{0.0, 0.0}, ip, 8080, *blob, bszone)
)

func Test_NewNode(t *testing.T) {
	fmt.Println("=========================Node Initialization Start===========================")

	blob := NewBlobInfo("A", 0, 0, 0)
	ip, ipnet, err := net.ParseCIDR("127.0.0.1/32")
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Println("-> ip:", ip, " net:", ipnet)
	ZoneA := NewZoneDefault()
	NodeA := NewDhtNode(Coordinates{0.5, 0.5}, ip, 8080, *blob, ZoneA)
	fmt.Println("mynode:", NodeA.blobInfo.playerName)
	i := strings.Compare(NodeA.blobInfo.playerName, "A")
	if i != 0 {
		t.Fatalf("Expected %s, got %s.", "A", NodeA.blobInfo.playerName)
	}
	fmt.Println("=========================Node Initialization Finish==========================")

}

func Test_NewBoard(t *testing.T) {
	fmt.Println("========================Keyspace Initialization Start========================")
	newPlayer := "newPlayer"
	keyspace := newKeySpace(bootstrapNode, newPlayer, "127.0.0.6/32", 8088)
	//make(keyspace.neighbors, nil)
	i := strings.Compare(keyspace.self.blobInfo.playerName, newPlayer)
	if i != 0 {
		t.Fatalf("Expected %s, got %s.", "Bootstrap", keyspace.self.blobInfo.playerName)
	}

	// The Bootstrap node should have node id {0,0}
	bid := keyspace.BootstrapNode.nodeId
	if bid.x != 0.0 || bid.y != 0.0 {
		t.Fatalf("Expected (0.0,0.0) got (%g,%g).", bid.x, bid.y)
	}
	sid := keyspace.self.nodeId
	fmt.Println("self has node id:", sid.x, sid.y)

	// Initialize 2 neighbors for the bootstrap node
	blob2 := NewBlobInfo("B", 0, 0, 0)
	ip2, _, _ := net.ParseCIDR("127.0.0.2/32")
	ZoneB := NewZoneDefault()
	NodeB := NewDhtNode(Coordinates{0.0, 0.5}, ip2, 8082, *blob2, ZoneB)

	blob3 := NewBlobInfo("C", 0, 0, 0)
	ip3, _, _ := net.ParseCIDR("127.0.0.3/32")
	ZoneC := NewZoneDefault()
	NodeC := NewDhtNode(Coordinates{0.5, 0.0}, ip3, 8083, *blob3, ZoneC)

	keyspace.neighbors[0] = NodeB
	keyspace.neighbors[1] = NodeC

	fmt.Printf("Neighbors for %s\n", keyspace.self.blobInfo.playerName)
	for i = 0; i < 8; i++ {
		if keyspace.neighbors[i] != nil {
			fmt.Println(keyspace.neighbors[i].blobInfo.playerName)
		}
	}
	fmt.Println("========================Keyspace Initialization Finish=======================")

}

func Test_Routing(t *testing.T) {
	rand.Seed(43)
	// The first node join
	newPlayer := "Player1"
	keyspace := newKeySpace(bootstrapNode, newPlayer, "127.0.0.6/32", 8088)
	fmt.Printf("The randomly selected point for the newly created keyspace is (%.2f, %.2f)",
		keyspace.self.nodeId.x, keyspace.self.nodeId.y)
	next := keyspace.RoutetoPoint(&keyspace.self.nodeId, keyspace.BootstrapNode)
	if strings.Compare(next.blobInfo.playerName, "Bootstrap") != 0 {
		t.Fatalf("Wrong owner: [%s] returned, should be [Bootstrap]", next.blobInfo.playerName)
	}
	fmt.Printf("next node returned by the routing dunction is %s", next.blobInfo.playerName)
	newZone, new_neighbors := next.splitZone(keyspace.self)
	fmt.Printf("New player's zone is {(%.2f, %.2f), (%.2f, %.2f)}\n", newZone.ul.x, newZone.ul.y, newZone.lr.x, newZone.lr.y)
	keyspace.self.zone = newZone
	fmt.Println(new_neighbors[0].blobInfo.playerName)

	rand.Seed(12)
	newPlayer2 := "Player2"
	keyspace2 := newKeySpace(bootstrapNode, newPlayer2, "127.0.0.7/32", 8089)
	fmt.Printf("The randomly selected point for the newly created keyspace is (%.2f, %.2f)",
		keyspace2.self.nodeId.x, keyspace2.self.nodeId.y)
	//keyspace2.self.zone.lr = Coordinates{}

}

func Test_SplitZone(t *testing.T) {
	newPlayer := "Player1"
	keyspace := newKeySpace(bootstrapNode, newPlayer, "127.0.0.6/32", 8088)
	fmt.Printf("The randomly selected point for the newly created keyspace is (%.2f, %.2f)",
		keyspace.self.nodeId.x, keyspace.self.nodeId.y)
	newZone, new_neighbors := bootstrapNode.splitZone(keyspace.self)
	fmt.Printf("New player's zone is {(%.2f, %.2f), (%.2f, %.2f)}\n", newZone.ul.x, newZone.ul.y, newZone.lr.x, newZone.lr.y)
	fmt.Printf("Old player's zone is {(%.2f, %.2f), (%.2f, %.2f)}\n",
		bootstrapNode.zone.ul.x, bootstrapNode.zone.ul.y, bootstrapNode.zone.lr.x, bootstrapNode.zone.lr.y)
	fmt.Println("new neighbor is:")
	fmt.Println(new_neighbors[0].blobInfo.playerName)

}
