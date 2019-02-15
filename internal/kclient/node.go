package kclient

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
	"time"
)

// Node represents a k8s node
type Node struct {
	// Address is the IPv4 address of the node.
	Address string
	// Ready is true when node can receive pods.
	Ready bool
	// LastSeen time
	LastSeen time.Time

	// Units contain the desired state of a node.
	Units map[string]string

	/* Status maintained by back-end */

	//
	LastReconcile time.Time
	//
	LastReconcileSuccess bool


	// Private state
	// Resource is the k8s Node that is changed.
	resource runtime.Object
}

// String returns a human readable representation of the receiver.
func (no *Node) String() string {
	var ss []string
	for s,_ := range no.Units {
		ss = append(ss, s)
	}
	return fmt.Sprintf("%s ready=%t units=%s", no.Address, no.Ready, strings.Join(ss,","))
}

