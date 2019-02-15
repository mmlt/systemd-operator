package kclient

import (
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"strings"
	"sync"
	"time"
)

// Controller watches Kubernetes ConfigMaps and Nodes resources.
type kclient struct {
	// operatorId is a string that is added to systemd unit names so they can be identified as managed by this operator.
	operatorId string
	// Client for the k8s API Server
	client kubernetes.Interface
	// Recorder to provide user feedback via Events.
	recorder record.EventRecorder
	// *StoreSynced func returns true when configMapStore is in sync with the API Server.
	configMapStoreSynced cache.InformerSynced
	nodeStoreSynced      cache.InformerSynced

	// changes is a worker queue that buffers the changes before they are send to the back-end via the OnChange supplied function.
	changes *changeQueue

	// nodes contains the nodes found in the cluster.
	nodes map[string]*Node
	// units
	units map[string]string
}

// StoreToConfigMapLister makes a Store that lists ConfigMap.
type StoreToConfigMapLister struct {
	cache.Store
}

const (
	// operatorName is shown in Event 'From' field.
	operatorName = "nto"
	// configLabelKey and Value are used to select the ConfigMaps seen by this controller.
	configLabelKey   = "operator"
	configLabelValue = "nto"
)

// New creates an API server client and subscribes to resource changes.
func New(kubeclientset kubernetes.Interface, sharedInformers informers.SharedInformerFactory, operatorId string) *kclient {
	// create event recorder
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: operatorName})

	// create informers
	configMapInformer := sharedInformers.Core().V1().ConfigMaps()
	nodeInformer := sharedInformers.Core().V1().Nodes()

	c := kclient{
		operatorId:           operatorId,
		client:               kubeclientset,
		recorder:             recorder,
		configMapStoreSynced: configMapInformer.Informer().HasSynced,
		nodeStoreSynced:      nodeInformer.Informer().HasSynced,

		nodes: make(map[string]*Node),
		units: make(map[string]string),
	}

	// queue that invokes backend function to process changes.
	c.changes = NewChangeQueue(nil)

	// ConfigMap changes
	configMapInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				c.configMapChange(Add, obj.(*corev1.ConfigMap))
			},
			UpdateFunc: func(old, cur interface{}) {
				c.configMapChange(Update, cur.(*corev1.ConfigMap))
			},
			DeleteFunc: func(obj interface{}) {
				c.configMapChange(Delete, obj.(*corev1.ConfigMap))
			},
		},
	)

	// Node changes
	nodeInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				c.nodeChange(Add, obj.(*corev1.Node))
			},
			UpdateFunc: func(old, cur interface{}) {
				c.nodeChange(Update, cur.(*corev1.Node))
			},
			DeleteFunc: func(obj interface{}) {
				c.nodeChange(Delete, obj.(*corev1.Node))
			},
		},
	)
	return &c
}

// OnChange sets the method that will be called when a Instruction is detected.
func (kc *kclient) OnChange(fn func(*Instruction)) {
	kc.changes.OnWork(fn)
}

// Start the client.
func (kc *kclient) Run(stopCh chan struct{}, wg *sync.WaitGroup) {
	go kc.changes.run(stopCh, wg)
}

// Event receives events and forwards them to the k8s API server so the user knows what's happening.
// 'resource' is the k8s resource that causes this event
func (kc *kclient) Event(node *Node, eventType, reason, message string) {
	kc.recorder.Eventf(node.resource, eventType, reason, message)
}

/***** API Instruction handlers ****************************************************/

// ConfigMapChange
func (kc *kclient) configMapChange(op OpCode, apiConfigMap *corev1.ConfigMap) {
	// only interested in ConfigMaps with label configLabelKey=configLabelValue
	// (a more efficient way is to use options.LabelSelector = labels.Set{configLabelKey: configLabelValue}.AsSelector() )
	v, ok := apiConfigMap.ObjectMeta.Labels[configLabelKey]
	if !ok || v != configLabelValue {
		return
	}

	glog.V(7).Infof("configMapChange %s %#v", op, apiConfigMap)

	switch op {
	case Add, Update:
		// get units from ConfigMap
		kc.units = make(map[string]string, len(apiConfigMap.Data))
		for k, v := range apiConfigMap.Data {
			kc.units[k] = strings.TrimSpace(v)
		}
	case Delete:
		kc.units = nil
	}

	// queue instructions to visit all nodes
	for _, v := range kc.nodes {
		v.Units = kc.units
		kc.changes.enqueue(&Instruction{op, v})
	}
}

// NodeChange updates the local list of nodes and optionally pushes a change notification
func (kc *kclient) nodeChange(op OpCode, apiNode *corev1.Node) {
	glog.V(7).Infof("nodeChange %v %v", op, apiNode)

	if kc.configMapStoreSynced() == false {
		// Ignore node changes as long as ConfigMaps aren't sync'd.
		// This is possible because node changes are send frequently.
		return
	}

	address := apiNode.Status.Addresses[0].Address
	n, ok := kc.nodes[address]
	if !ok {
		n = &Node{
			Address: address,
		}
		kc.nodes[address] = n
	}

	var ready bool
	switch op {
		case Add, Update:
			for _, c := range apiNode.Status.Conditions {
				if c.Type == corev1.NodeReady {
					ready = (c.Status == corev1.ConditionTrue)
					break
				}
			}
			n.Units = kc.units
			n.LastSeen = time.Now()
		case Delete:
			ready = false
			n.Units = nil
	}

	if n.Ready != ready {
		n.Ready = ready
		kc.changes.enqueue(&Instruction{op, n})
	}
}

