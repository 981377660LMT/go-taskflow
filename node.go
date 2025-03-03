package gotaskflow

import (
	"sync"
	"sync/atomic"

	"github.com/noneback/go-taskflow/utils"
)

const (
	kNodeStateIdle     = int32(0)
	kNodeStateWaiting  = int32(1)
	kNodeStateRunning  = int32(2)
	kNodeStateFinished = int32(3)
	kNodeStateFailed   = int32(4)
	// kNodeStateCanceled = int32(5)
)

type nodeType string

const (
	nodeSubflow   nodeType = "subflow"   // subflow
	nodeStatic    nodeType = "static"    // static
	nodeCondition nodeType = "condition" // static
)

type innerNode struct {
	name        string
	successors  []*innerNode
	dependents  []*innerNode
	Typ         nodeType
	ptr         interface{} // 存储具体任务实现
	rw          *sync.RWMutex
	state       atomic.Int32 // 任务状态
	joinCounter *utils.RC    // 入度计数器
	g           *eGraph
	priority    TaskPriority
}

func (n *innerNode) JoinCounter() int {
	return n.joinCounter.Value()
}

func (n *innerNode) setup() {
	n.state.Store(kNodeStateIdle)
	for _, dep := range n.dependents {
		if dep.Typ == nodeCondition {
			continue
		}

		n.joinCounter.Increase()
	}
}

func (n *innerNode) drop() {
	// release every deps
	for _, node := range n.successors {
		if n.Typ != nodeCondition {
			node.joinCounter.Decrease()
		}
	}
}

// set dependency： V deps on N, V is input node
func (n *innerNode) precede(v *innerNode) {
	n.successors = append(n.successors, v)
	v.dependents = append(v.dependents, n)
}

func newNode(name string) *innerNode {
	return &innerNode{
		name:        name,
		state:       atomic.Int32{},
		successors:  make([]*innerNode, 0),
		dependents:  make([]*innerNode, 0),
		rw:          &sync.RWMutex{},
		priority:    NORMAL,
		joinCounter: utils.NewRC(),
	}
}
