package strategies

import (
	"container/heap"
	"github.com/goletan/services-library/shared/types"
)

const DefaultServicePriority = 99

// ServiceEndpointItem represents an item in the priority queue.
type ServiceEndpointItem struct {
	Priority int
	Endpoint types.ServiceEndpoint
}

// PriorityQueue manages service endpoint prioritization.
type PriorityQueue []*ServiceEndpointItem

// Methods required by the heap.Interface for PriorityQueue.

// Len returns the number of items currently in the priority queue.
func (pq *PriorityQueue) Len() int { return len(*pq) }

// Less compares the priorities of two items in the priority queue and returns true if the first has a lower priority.
func (pq *PriorityQueue) Less(i, j int) bool { return (*pq)[i].Priority < (*pq)[j].Priority }

// Swap exchanges the elements at indices i and j in the priority queue.
func (pq *PriorityQueue) Swap(i, j int) { (*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i] }

// Push adds an item to the queue.
func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*ServiceEndpointItem))
}

// Pop removes and returns the highest-priority item.
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// PriorityQueueManager handles the queue logic and priority mapping.
type PriorityQueueManager struct {
	queue       PriorityQueue
	priorityMap map[string]int
}

// NewPriorityQueueManager initializes a new manager.
func NewPriorityQueueManager(priorityMap map[string]int) *PriorityQueueManager {
	return &PriorityQueueManager{
		queue:       make(PriorityQueue, 0),
		priorityMap: priorityMap,
	}
}

// Push adds a service to the queue either at the mapped priority or default.
func (pqm *PriorityQueueManager) Push(endpoint types.ServiceEndpoint) {
	priority := pqm.priorityMap[endpoint.Name]
	if priority == 0 {
		priority = DefaultServicePriority
	}
	heap.Push(&pqm.queue, &ServiceEndpointItem{Priority: priority, Endpoint: endpoint})
}

// Pop retrieves the highest-priority service item.
func (pqm *PriorityQueueManager) Pop() *ServiceEndpointItem {
	return heap.Pop(&pqm.queue).(*ServiceEndpointItem)
}

// Len returns the number of items in the queue.
func (pqm *PriorityQueueManager) Len() int {
	return pqm.queue.Len()
}
