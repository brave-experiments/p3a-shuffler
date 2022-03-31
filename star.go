package main

import (
	"fmt"
	"sync"
)

const (
	orderHighEntropyFirst = iota
	orderHighEntropyLast
)

// NestedSTAR simulates the execution of Nested STAR over P3A measurements.
// Note that this struct doesn't actually implement Nested STAR; it merely
// simulates its nesting (represented as a tree of nodes) to produce CSV output
// that allows us to explore the privacy and utility tradeoff.
type NestedSTAR struct {
	sync.WaitGroup
	inbox           chan []Report
	root            *Node
	threshold       int
	order           int
	numMeasurements int
}

// NewNestedSTAR returns a new NestedSTAR object.
func NewNestedSTAR(cfg *simulationConfig) *NestedSTAR {
	return &NestedSTAR{
		inbox:     make(chan []Report),
		root:      &Node{make(map[string]*NodeInfo)},
		threshold: cfg.AnonymityThreshold,
		order:     cfg.Order,
	}
}

// AddReports adds the given reports to Nested STAR.  The argument 'method'
// refers to the subset of attributes that we consider.
func (s *NestedSTAR) AddReports(method int, reports []Report) {
	s.numMeasurements += len(reports)
	var m P3AMeasurement
	for _, r := range reports {
		m = r.(P3AMeasurement)
		if s.order == orderHighEntropyFirst {
			s.root.Add(m.OrderHighEntropyFirst(method))
		} else {
			s.root.Add(m.OrderHighEntropyLast(method))
		}
	}
}

func frac(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b)
}

// Aggregate aggregates Nested STAR's measurements.  The argument 'method'
// refers to the subset of attributes we consider and 'numAttrs' refers to the
// number of attributes.
func (s *NestedSTAR) Aggregate(method int, numAttrs int) {
	state := s.root.Aggregate(numAttrs, s.threshold, []string{})
	for key := 1; key <= numAttrs; key++ {
		num, exists := state.LenPartialMsmts[key]
		if !exists {
			num = 0
		}
		fmt.Printf("LenPartMsmt%s,%d,%d,0,0,0,%d,%d\n",
			anonymityAttrs[method],
			s.order,
			s.threshold,
			key,
			num)
	}
	fracFull := frac(state.FullMsmts, s.numMeasurements) * 100
	fracPart := frac(state.PartialMsmts, s.numMeasurements) * 100
	elog.Printf("%d (%.1f%%) full, %d (%.1f%%) partial out of %d; %.1f%% lost\n",
		state.FullMsmts,
		fracFull,
		state.PartialMsmts,
		fracPart,
		s.numMeasurements,
		100-fracFull-fracPart)
	fmt.Printf("Partial%s,%d,%d,%.3f,%d,%d,0,0\n",
		anonymityAttrs[method],
		s.order,
		s.threshold,
		frac(state.PartialMsmts, s.numMeasurements),
		s.root.NumTags(),
		s.root.NumLeafTags())
}

type NodeInfo struct {
	Num  int
	Next *Node
}

type Node struct {
	// E.g., "US"    -> NodeInfo{}
	//       "Linux" -> NodeInfo{}
	ValueToInfo map[string]*NodeInfo
}

type AggregationState struct {
	FullMsmts       int
	PartialMsmts    int
	LenPartialMsmts map[int]int
}

func (s *AggregationState) String() string {
	return fmt.Sprintf("%d full, %d partial msmts.", s.FullMsmts, s.PartialMsmts)
}

func (s *AggregationState) AddLenTags(key, value int) {
	num, exists := s.LenPartialMsmts[key]
	if !exists {
		s.LenPartialMsmts[key] = value
	} else {
		s.LenPartialMsmts[key] = num + value
	}
}

func (s *AggregationState) Augment(s2 *AggregationState) {
	s.FullMsmts += s2.FullMsmts
	s.PartialMsmts += s2.PartialMsmts
	for key, value := range s2.LenPartialMsmts {
		s.AddLenTags(key, value)
	}
}

func (s *AggregationState) NothingUnlocked() bool {
	return s.FullMsmts == 0 && s.PartialMsmts == 0
}

func NewAggregationState() *AggregationState {
	return &AggregationState{
		LenPartialMsmts: make(map[int]int),
	}
}

func (n *Node) Aggregate(maxTags, threshold int, m []string) *AggregationState {
	state := NewAggregationState()

	// Iterate over all values where we are in the tree, e.g., "US", "FR", ...
	for value, info := range n.ValueToInfo {
		// We don't meet our k-anonymity threshold for the given value.
		if info.Num < threshold {
			continue
		}

		// We've reached the last tag, i.e., we fully unlocked a measurement.
		if len(m)+1 == maxTags {
			state.FullMsmts += info.Num
			continue
		}

		if info.Next == nil {
			// This branch is only entered if we're dealing with incomplete
			// measurements.
			state.PartialMsmts += info.Num
			state.AddLenTags(len(m)+1, info.Num)
		} else {
			// Go deeper down the tree, and try to unlock our next tag.
			subState := info.Next.Aggregate(maxTags, threshold, append(m, value))
			state.Augment(subState)
			state.PartialMsmts = info.Num - subState.FullMsmts
			state.AddLenTags(len(m)+1, state.PartialMsmts-subState.PartialMsmts)
		}
	}

	return state
}

func (n *Node) Add(orderedMsmt []string) {
	info, exists := n.ValueToInfo[orderedMsmt[0]]
	if !exists {
		info = &NodeInfo{Num: 1}
		n.ValueToInfo[orderedMsmt[0]] = info
	} else {
		info.Num++
	}

	if len(orderedMsmt[1:]) > 0 {
		if info.Next == nil {
			newNode := &Node{ValueToInfo: make(map[string]*NodeInfo)}
			info.Next = newNode
			newNode.Add(orderedMsmt[1:])
		} else {
			info.Next.Add(orderedMsmt[1:])
		}
	}
}

func (n *Node) NumTags() int {
	var num = len(n.ValueToInfo)

	for _, info := range n.ValueToInfo {
		if info.Next != nil {
			num += info.Next.NumTags()
		}
	}
	return num
}

func (n *Node) NumNodes() int {
	var num = 1

	for _, info := range n.ValueToInfo {
		if info.Next != nil {
			num += info.Next.NumNodes()
		}
	}
	return num
}

func (n *Node) NumLeafTags() int {
	var num int

	for _, info := range n.ValueToInfo {
		if info.Next == nil {
			num++
		} else {
			num += info.Next.NumLeafTags()
		}
	}
	return num
}
