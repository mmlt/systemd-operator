/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kclient

import (
	//"k8s.io/kubernetes/pkg/util/workqueue"
	"k8s.io/client-go/util/workqueue"
	"sync"
)

// changeQueue manages a queue with a worker function.
type changeQueue struct {
	// queue is the work queue the worker polls
	queue *workqueue.Type
	// workFn is called for each item in the queue.
	workFn func(*Instruction)
}

// NewChangeQueue creates a queue with a function that's called for every enqueued Instruction.
func NewChangeQueue(fn func(*Instruction)) *changeQueue {
	return &changeQueue{
		queue:  workqueue.New(),
		workFn: fn,
	}
}

// OnWork sets the function that consumes queued items.
func (t *changeQueue) OnWork(fn func(*Instruction)) {
	t.workFn = fn
}

// run the worker function until the stopCh is closed.
func (t *changeQueue) run(stopCh <-chan struct{}, wg *sync.WaitGroup) {
	wg.Add(1)
	go t.work()
	select {
	case <-stopCh:
		t.queue.ShutDown()
		// note that we don't wait for t.queue.ShuttingDown() to become false (for the queue to drain).
		wg.Done()
		return
	}
}

// enqueue a change for the worker function to process.
func (t *changeQueue) enqueue(ch *Instruction) { //TODO rename to add()
	t.queue.Add(ch)
}

// work gets an item from the queue and runs the workFn.
func (t *changeQueue) work() {
	for {
		change, quit := t.queue.Get()
		if quit {
			return
		}
		change2, ok := change.(*Instruction)
		if ok && t.workFn != nil {
			t.workFn(change2)
		}
		t.queue.Done(change)
	}
}
