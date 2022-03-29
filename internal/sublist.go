// Copyright 2012 Apcera Inc. All rights reserved.

package internal

import (
	"sync"
	"sync/atomic"
	"time"
)

// A Sublist stores and efficiently retrieves subscriptions. It uses a
// tree structure and an efficient RR cache to achieve quick lookups.
type Sublist struct {
	mu    sync.RWMutex
	root  *level
	count uint32
	cache map[string][]interface{}
	cmax  int
	stats stats
}

type stats struct {
	inserts   uint64
	removes   uint64
	matches   uint64
	cacheHits uint64
	since     time.Time
}

// A node contains subscriptions and a pointer to the next level.
type node struct {
	next *level
	subs []interface{}
}

// A level represents a group of nodes and special pointers to
// wildcard nodes.
type level struct {
	nodes    map[string]*node
	pwc, fwc *node
}

// Create a new default node.
func newNode() *node {
	return &node{subs: make([]interface{}, 0, 4)}
}

// Create a new default level. We use FNV1A as the hash
// algortihm for the tokens, which should be short.
func newLevel() *level {
	return &level{nodes: make(map[string]*node)}
}

// defaultCacheMax is used to bound limit the frontend cache
const defaultCacheMax = 1024

// NewSublist will create a default sublist
func NewSublist() *Sublist {
	return &Sublist{
		root:  newLevel(),
		cache: make(map[string][]interface{}),
		cmax:  defaultCacheMax,
		stats: stats{since: time.Now()},
	}
}

// Common byte variables for wildcards and token separator.
var (
	_PWC = byte('*')
	_FWC = byte('>')
	_SEP = byte('.')
)

// split will split a subject into tokens
func split(subject string, tokens []string) []string {
	start := 0
	for i := range subject {
		if subject[i] == _SEP {
			tokens = append(tokens, subject[start:i])
			start = i + 1
		}
	}
	return append(tokens, subject[start:])
}

func (s *Sublist) Insert(subject string, sub interface{}) {
	tsa := [16]string{}
	toks := split(subject, tsa[:0])

	s.mu.Lock()
	l := s.root
	var n *node

	for _, t := range toks {
		switch t[0] {
		case _PWC:
			n = l.pwc
		case _FWC:
			n = l.fwc
		default:
			n = l.nodes[t]
		}
		if n == nil {
			n = newNode()
			switch t[0] {
			case _PWC:
				l.pwc = n
			case _FWC:
				l.fwc = n
			default:
				l.nodes[t] = n
			}
		}
		if n.next == nil {
			n.next = newLevel()
		}
		l = n.next
	}
	n.subs = append(n.subs, sub)
	s.count++
	s.stats.inserts++
	s.addToCache(subject, sub)
	s.mu.Unlock()
}

// addToCache will add the new entry to existing cache
// entries if needed.
func (s *Sublist) addToCache(subject string, sub interface{}) {
	if len(s.cache) == 0 {
		return
	}
	for k := range s.cache {
		if !matchLiteral(k, subject) {
			continue
		}

		if s.cache[k] == nil {
			continue
		}
		s.cache[k] = append(s.cache[k], sub)
	}
}

// removeFromCache will remove the sub from any active cache entries
func (s *Sublist) removeFromCache(subject string, sub interface{}) {
	if len(s.cache) == 0 {
		return
	}
	for k := range s.cache {
		if !matchLiteral(k, subject) {
			continue
		}
		delete(s.cache, k)
		sub = s.cache[k]
		delete(s.cache, k)
	}
}

// Match will match all entries to the literal subject. It will return a
// slice of results.
func (s *Sublist) Match(subject string) []interface{} {
	s.mu.RLock()
	atomic.AddUint64(&s.stats.matches, 1)
	r := s.cache[subject]
	s.mu.RUnlock()

	if r != nil {
		atomic.AddUint64(&s.stats.cacheHits, 1)
		return r
	}

	// Cache miss
	// Process subject into tokens, this is performed
	// unlocked, so can be parallel.
	tsa := [32]string{}
	toks := tsa[:0]

	start := 0
	for i := range subject {
		if subject[i] == _SEP {
			toks = append(toks, subject[start:i])
			start = i + 1
		}
	}
	toks = append(toks, subject[start:])
	results := make([]interface{}, 0, 4)

	// Lookup and add entry to hash.
	s.mu.Lock()
	matchLevel(s.root, toks, &results)

	// We use random eviction to bound the size of the cache.
	// RR is used for speed purposes here.
	if len(s.cache) >= s.cmax {
		// s.cache.RemoveRandom() TODO
	}
	s.cache[string(subject)] = results
	s.mu.Unlock()

	return results
}

// matchLevel is used to recursively descend into the trie when there
// is a cache miss.
func matchLevel(l *level, toks []string, results *[]interface{}) {
	var pwc, n *node
	for i, t := range toks {
		if l == nil {
			return
		}
		if l.fwc != nil {
			*results = append(*results, l.fwc.subs...)
		}
		if pwc = l.pwc; pwc != nil {
			matchLevel(pwc.next, toks[i+1:], results)
		}

		n = l.nodes[t]
		if n != nil {
			l = n.next
		} else {
			l = nil
		}
	}
	if n != nil {
		*results = append(*results, n.subs...)
	}
	if pwc != nil {
		*results = append(*results, pwc.subs...)
	}
	return
}

// lnt is used to track descent into a removal for pruning.
type lnt struct {
	l *level
	n *node
	t string
}

// Remove will remove any item associated with key. It will track descent
// into the trie and prune upon successful removal.
func (s *Sublist) Remove(subject string, sub interface{}) {
	tsa := [16]string{}
	toks := split(subject, tsa[:0])

	s.mu.Lock()
	l := s.root
	var n *node

	var lnts [32]lnt
	levels := lnts[:0]

	for _, t := range toks {
		if l == nil {
			s.mu.Unlock()
			return
		}
		switch t[0] {
		case _PWC:
			n = l.pwc
		case _FWC:
			n = l.fwc
		default:
			n = l.nodes[string(t)]
		}
		if n != nil {
			levels = append(levels, lnt{l, n, t})
			l = n.next
		} else {
			l = nil
		}
	}
	if !s.removeFromNode(n, sub) {
		s.mu.Unlock()
		return
	}

	s.count--
	s.stats.removes++

	for i := len(levels) - 1; i >= 0; i-- {
		l, n, t := levels[i].l, levels[i].n, levels[i].t
		if n.isEmpty() {
			l.pruneNode(n, t)
		}
	}
	s.removeFromCache(subject, sub)
	s.mu.Unlock()
}

// pruneNode is used to prune and empty node from the tree.
func (l *level) pruneNode(n *node, t string) {
	if n == nil {
		return
	}
	if n == l.fwc {
		l.fwc = nil
	} else if n == l.pwc {
		l.pwc = nil
	} else {
		delete(l.nodes, t)
	}
}

// isEmpty will test if the node has any entries. Used
// in pruning.
func (n *node) isEmpty() bool {
	if len(n.subs) == 0 {
		if n.next == nil || n.next.numNodes() == 0 {
			return true
		}
	}
	return false
}

// Return the number of nodes for the given level.
func (l *level) numNodes() uint32 {
	num := len(l.nodes)
	if l.pwc != nil {
		num += 1
	}
	if l.fwc != nil {
		num += 1
	}
	return uint32(num)
}

// Remove the sub for the given node.
func (s *Sublist) removeFromNode(n *node, sub interface{}) bool {
	if n == nil {
		return false
	}
	for i, v := range n.subs {
		if v == sub {
			num := len(n.subs)
			a := n.subs
			copy(a[i:num-1], a[i+1:num])
			n.subs = a[0 : num-1]
			return true
		}
	}
	return false
}

// matchLiteral is used to test literal subjects, those that do not have any
// wildcards, with a target subject. This is used in the cache layer.
func matchLiteral(literal, subject string) bool {
	li := 0
	for i := range subject {
		b := subject[i]
		if li >= len(literal) {
			return false
		}
		switch b {
		case _PWC:
			// Skip token in literal
			ll := len(literal)
			for {
				if li >= ll || literal[li] == _SEP {
					li -= 1
					break
				}
				li += 1
			}
		case _FWC:
			return true
		default:
			if b != literal[li] {
				return false
			}
		}
		li += 1
	}
	return true
}

// Count return the number of stored items in the HashMap.
func (s *Sublist) Count() uint32 { return s.count }

// Stats for the sublist
type Stats struct {
	NumSubs      uint32
	NumCache     uint32
	NumInserts   uint64
	NumRemoves   uint64
	NumMatches   uint64
	CacheHitRate float64
	MaxFanout    uint32
	AvgFanout    float64
	StatsTime    time.Time
}

// Stats will return a stats structure for the current state.
func (s *Sublist) Stats() *Stats {
	s.mu.Lock()
	defer s.mu.Unlock()

	st := &Stats{}
	st.NumSubs = s.count
	st.NumCache = uint32(len(s.cache))
	st.NumInserts = s.stats.inserts
	st.NumRemoves = s.stats.removes
	st.NumMatches = s.stats.matches
	if s.stats.matches > 0 {
		st.CacheHitRate = float64(s.stats.cacheHits) / float64(s.stats.matches)
	}
	// whip through cache for fanout stats
	// FIXME, creating all each time could be expensive, should do a cb version.
	tot, max := 0, 0
	for _, r := range s.cache {
		l := len(r)
		tot += l
		if l > max {
			max = l
		}
	}
	st.MaxFanout = uint32(max)
	st.AvgFanout = float64(tot) / float64(len(s.cache))
	st.StatsTime = s.stats.since
	return st
}

// ResetStats will clear stats and update StatsTime to time.Now()
func (s *Sublist) ResetStats() {
	s.stats = stats{}
	s.stats.since = time.Now()
}

// numLevels will return the maximum number of levels
// contained in the Sublist tree.
func (s *Sublist) numLevels() int {
	return visitLevel(s.root, 0)
}

// visitLevel is used to descend the Sublist tree structure
// recursively.
func visitLevel(l *level, depth int) int {
	if l == nil || l.numNodes() == 0 {
		return depth
	}

	depth += 1
	maxDepth := depth

	for _, n := range l.nodes {
		if n == nil {
			continue
		}
		newDepth := visitLevel(n.next, depth)
		if newDepth > maxDepth {
			maxDepth = newDepth
		}
	}
	if l.pwc != nil {
		pwcDepth := visitLevel(l.pwc.next, depth)
		if pwcDepth > maxDepth {
			maxDepth = pwcDepth
		}
	}
	if l.fwc != nil {
		fwcDepth := visitLevel(l.fwc.next, depth)
		if fwcDepth > maxDepth {
			maxDepth = fwcDepth
		}
	}
	return maxDepth
}
