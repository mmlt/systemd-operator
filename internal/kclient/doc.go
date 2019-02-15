// Package kclient is responsible for maintaining 'desired state' of the operator.
package kclient


// Package kclient is responsible for maintaining 'desired state' of the operator.
// It watches the Kubernetes API Server for changes in ConfigMaps and Nodes and updates it's Nodes and Timers
// data structure accordingly.
// On changes it enqueues an OpCode to be performed by the backend.


