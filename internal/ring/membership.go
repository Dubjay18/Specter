package ring

import (
	"log"
	"net"
	"strconv"

	"github.com/hashicorp/memberlist"
)

type EventDelegate struct {
	ring *Ring
}


func NewEventDelegate(ring *Ring) *EventDelegate {
	return &EventDelegate{ring: ring}
}

func (d *EventDelegate) NotifyJoin(node *memberlist.Node) {
	 log.Printf("ring: node joined → %s", node.Name)
	d.ring.AddNode(node.Name)
}

func (d *EventDelegate) NotifyLeave(node *memberlist.Node) {
	 log.Printf("ring: node left → %s", node.Name)
	d.ring.RemoveNode(node.Name)
}

func (d *EventDelegate) NotifyUpdate(node *memberlist.Node) {
	// No action needed for updates in this simple implementation
}

func StartMembership(nodeName, bindAddr string, peers []string, ring *Ring) (*memberlist.Memberlist, error){
	config := memberlist.DefaultLocalConfig()
	eventDelegate := &EventDelegate{ring: ring}
	config.Name = nodeName
	host, port, _ := net.SplitHostPort(bindAddr)
	config.BindAddr = host
	portNum, _ := strconv.Atoi(port)
	config.BindPort = portNum
	config.Events = eventDelegate

	ml, err := memberlist.Create(config)
	if err != nil {
		return nil, err
	}

	if len(peers) > 0 {
		_, err = ml.Join(peers)
		if err != nil {
			return nil, err
		}
	}

	return ml, nil
}