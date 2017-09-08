package CAN

import (
	"net"
	"strconv"
)

const (
	UP = iota
	DOWN
	LEFT
	RIGHT
)

type Coordinates struct {
	x, y float64
}

type Zone struct {
	ul, lr Coordinates
}

type BlobInfo struct {
	playerName string
	size       float64
	speed      float64
	direction  int
}

type DhtNode struct {
	nodeId    Coordinates
	ipAddr    net.IP
	port      int
	blobInfo  BlobInfo
	zone      *Zone
	neighbors [8]*DhtNode
}

func (node *DhtNode) String() string {
	x_str := strconv.FormatFloat(node.nodeId.x, 'f', -1, 64)
	y_str := strconv.FormatFloat(node.nodeId.y, 'f', -1, 64)
	return "Node<" + x_str + "," + y_str + ">"
	// return "Node<" + node.ipAddr.String() + ":" +
	// 	string(node.port) + "/" + node.blobInfo.playerName;
}

func NewZone(ulc Coordinates, lrc Coordinates) *Zone {
	return &Zone{
		ul: ulc,
		lr: lrc,
	}
}

func NewZoneDefault() *Zone {
	return &Zone{
		ul: Coordinates{-1, -1},
		lr: Coordinates{-1, -1},
	}
}

func NewDhtNode(id Coordinates, ip net.IP, p int, blob BlobInfo, z *Zone) *DhtNode {
	return &DhtNode{
		nodeId:    id,
		ipAddr:    ip,
		port:      p,
		blobInfo:  blob,
		zone:      z,
		neighbors: [8]*DhtNode{nil, nil, nil, nil, nil, nil, nil, nil},
	}
}

func NewBlobInfo(name string, s float64, sp float64, d int) *BlobInfo {
	return &BlobInfo{
		playerName: name,
		size:       s,
		speed:      sp,
		direction:  d,
	}
}

func (d *DhtNode) getNodeId() Coordinates {
	return d.nodeId
}

func (d *DhtNode) splitZone(sender *DhtNode) (*Zone, []*DhtNode) {
	x_d := d.zone.lr.x - d.zone.ul.x
	y_d := d.zone.lr.y - d.zone.ul.y
	new_neighbors := []*DhtNode{}
	z := sender.zone
	if x_d >= y_d {
		// split the zone along x-axis
		d.zone.lr.x = d.zone.ul.x + x_d/2
		z_ulc := Coordinates{d.zone.lr.x, d.zone.ul.y}
		z_lrc := Coordinates{d.zone.lr.x + x_d/2, d.zone.lr.y}
		z.lr = z_lrc
		z.ul = z_ulc
		for i := 0; i < 8; i++ {
			if d.neighbors[i] != nil {

				n_ul := d.neighbors[i].zone.ul
				//n_lr := k.self.neighbors[i].zone.lr
				if n_ul.x > d.zone.lr.x {
					new_neighbors = append(new_neighbors, d.neighbors[i])
				}
			}
		}
	} else {
		// split the zone along y-axis
		d.zone.lr.y = d.zone.ul.y + y_d/2
		z_ulc := Coordinates{d.zone.ul.x, d.zone.lr.y}
		z_lrc := Coordinates{d.zone.lr.x, d.zone.lr.y + y_d/2}
		z.lr = z_lrc
		z.ul = z_ulc
		for i := 0; i < 8; i++ {
			if d.neighbors[i] != nil {

				n_ul := d.neighbors[i].zone.ul
				//n_lr := k.self.neighbors[i].zone.lr
				if n_ul.y > d.zone.lr.y {
					new_neighbors = append(new_neighbors, d.neighbors[i])
				}
			}
		}
	}
	new_neighbors = append(new_neighbors, d)
	for i := 0; i < 8; i++ {
		if d.neighbors[i] == nil {
			d.neighbors[i] = sender
			break
		}
	}
	return z, new_neighbors
}
