package CAN

import (
	//"math"
	"fmt"
	"math/rand"
	"net"
)

type KeySpace struct {
	self          *DhtNode //bootstrap node
	BootstrapNode *DhtNode
	nodes         [4][4]*DhtNode // test with 4x4 board
	neighbors     [8]*DhtNode
}

type DhtRouting int

func Round(x, unit float64) float64 {
	return float64(int64(x/unit+0.5)) * unit
}

func newKeySpace(bsNode *DhtNode, player_name string, ipaddr string, port int) *KeySpace {
	// randomly select the nodeID in the keyspace
	x := Round(rand.Float64(), 0.01)
	y := Round(rand.Float64(), 0.01)
	newblob := NewBlobInfo(player_name, 0, 0, 0)
	ip, _, _ := net.ParseCIDR(ipaddr)
	newZone := NewZoneDefault()
	newNode := NewDhtNode(Coordinates{x, y}, ip, port, *newblob, newZone)
	return &KeySpace{
		self:          newNode,
		BootstrapNode: bsNode,
		nodes:         [4][4]*DhtNode{},
	}
}

func (k *KeySpace) RoutetoPoint(d *Coordinates, c *DhtNode) *DhtNode {
	// k is the newly created space, the node ID is randomly selected and
	// this function tries to find the owner of that point, and ask for a
	// share of the zone from the owner.
	// destination is the self nodeID, c is the bootstrap node
	var next *DhtNode
	if d.x < c.zone.lr.x && d.y < c.zone.lr.y {
		if d.x > c.zone.ul.x && d.y > c.zone.ul.y {
			return c
		}
	} else {
		dist := 2.0
		for i := 0; i < 8; i++ {
			n_x := c.neighbors[i].getNodeId().x
			n_y := c.neighbors[i].getNodeId().y
			if c.neighbors[i] != nil {
				curr_dist := (d.x-n_x)*(d.x-n_x) + (d.y-n_y)*(d.y-n_y)
				if curr_dist < dist {
					dist := curr_dist
					next := c.neighbors[i]
					fmt.Printf("current distance is %.6f from node: %s", dist, next.blobInfo.playerName)
				}
			}
		}
		return next
	}
	return next
}

func (k *KeySpace) splitZone(sender *DhtNode) (*Zone, []*DhtNode) {
	x_d := k.self.zone.lr.x - k.self.zone.ul.x
	y_d := k.self.zone.lr.y - k.self.zone.ul.y
	new_neighbors := []*DhtNode{}
	z := sender.zone
	if x_d >= y_d {
		// split the zone along x-axis
		k.self.zone.lr.x = k.self.zone.ul.x + x_d/2
		z_ulc := Coordinates{k.self.zone.lr.x, k.self.zone.ul.y}
		z_lrc := Coordinates{k.self.zone.lr.x + x_d/2, k.self.zone.lr.y}
		z.lr = z_lrc
		z.ul = z_ulc
		for i := 0; i < 8; i++ {
			if k.self.neighbors[i] != nil {

				n_ul := k.self.neighbors[i].zone.ul
				//n_lr := k.self.neighbors[i].zone.lr
				if n_ul.x > k.self.zone.lr.x {
					new_neighbors = append(new_neighbors, k.self.neighbors[i])
				}
			}
		}
	} else {
		// split the zone along y-axis
		k.self.zone.lr.y = k.self.zone.ul.y + y_d/2
		z_ulc := Coordinates{k.self.zone.ul.x, k.self.zone.lr.y}
		z_lrc := Coordinates{k.self.zone.lr.x, k.self.zone.lr.y + y_d/2}
		z.lr = z_lrc
		z.ul = z_ulc
		for i := 0; i < 8; i++ {
			if k.self.neighbors[i] != nil {

				n_ul := k.self.neighbors[i].zone.ul
				//n_lr := k.self.neighbors[i].zone.lr
				if n_ul.y > k.self.zone.lr.y {
					new_neighbors = append(new_neighbors, k.self.neighbors[i])
				}
			}
		}
	}
	new_neighbors = append(new_neighbors, k.self)
	for i := 0; i < 8; i++ {
		if k.self.neighbors[i] == nil {
			k.self.neighbors[i] = sender
			break
		}
	}
	return z, new_neighbors
}
