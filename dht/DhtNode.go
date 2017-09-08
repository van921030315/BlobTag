package dht

import (
	"net"
	"strconv"
	"time"
)

const NANOS_PER_MS = 1000000
const UI_FRAME_UPDATE_RATE_MS = 1000.0 / 60.0 //60 frames per sec
const VELOCITY_MS_RATE = 1.0 / UI_FRAME_UPDATE_RATE_MS
const (
	UP = iota
	DOWN
	LEFT
	RIGHT
)

type BlobInfo struct {
	PlayerName  string
	IsAlive     bool
	Color       int
	Xsize       float64
	Ysize       float64
	Xvel        float64
	Yvel        float64
	Xpos        float64
	Ypos        float64
	Zone        int
	ZoneChanged bool
}

type Coordinates struct {
	X, Y int
}

func (b *BlobInfo) String() string {
	return "B<" + b.PlayerName + ":(" + Ftoa(b.Xpos) + "," + Ftoa(b.Ypos) +
		"):(" + Ftoa(b.Xvel) + "," + Ftoa(b.Yvel) + ")>"
}

type DhtNode struct {
	NodeId   int
	IpAddr   net.IP
	Port     int
	BlobInfo *BlobInfo
	//time of last updating this struct (epoch time: Seconds since 1 Jan 00:00 1970)
	UpdateTimeNanos int64
}

func (node *DhtNode) String() string {
	return "Node<" + strconv.Itoa(node.NodeId) + ">:" + node.IpAddr.String() +
		":" + strconv.Itoa(node.Port) + "; Blob: " + node.BlobInfo.String() + ">"
}

/* Estimates blob's current position based on vx, vy and last updatetime */
func (node *DhtNode) UpdateBlobPositionEstimate() {
	timeNowMillis := time.Now().UnixNano() / NANOS_PER_MS
	updateTimeMillis := node.UpdateTimeNanos / NANOS_PER_MS

	timeDiffMs := float64(timeNowMillis - updateTimeMillis)
	node.BlobInfo.Xpos += timeDiffMs * node.BlobInfo.Xvel * VELOCITY_MS_RATE
	node.BlobInfo.Ypos += timeDiffMs * node.BlobInfo.Yvel * VELOCITY_MS_RATE
}

func Ftoa(fl float64) string {
	return strconv.FormatFloat(fl, 'f', -1, 64)
}

/* NodeInfo class */
type NodeInfo struct {
	NetAddr             string //eg "126.3.6.12:8000"
	NodeId              int
	LastUpdateEpochTime int64
	LastUpdateTimeStr   string
	IsActive            bool
}

/* NodeInfoHistory class */
type NodeInfoHistory struct {
	NodeMap map[int]*NodeInfo
}
