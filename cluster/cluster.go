package cluster

import (
//"net"
)

type Node interface {
	// ID returns a node identifier.
	ID() string
	Receive() error
	Send() error
}

type Cluster interface {
	Add(Node) error
	Del(Node) error
	Get(Node) (Node, error)
}
