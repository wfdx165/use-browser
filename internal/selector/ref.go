package selector

import (
	"fmt"
	"strconv"
	"strings"
)

type RefResolver struct {
	refToNode map[string]int64
	nodeToRef map[int64]string
}

func NewRefResolver() *RefResolver {
	return &RefResolver{
		refToNode: make(map[string]int64),
		nodeToRef: make(map[int64]string),
	}
}

func (r *RefResolver) SetMappings(mappings map[string]int64) {
	r.refToNode = make(map[string]int64, len(mappings))
	r.nodeToRef = make(map[int64]string, len(mappings))
	for ref, nodeID := range mappings {
		r.refToNode[ref] = nodeID
		r.nodeToRef[nodeID] = ref
	}
}

func (r *RefResolver) Resolve(ref string) (int64, bool) {
	if !strings.HasPrefix(ref, "@e") {
		return 0, false
	}
	id, ok := r.refToNode[ref]
	return id, ok
}

func (r *RefResolver) GetRef(nodeID int64) (string, bool) {
	ref, ok := r.nodeToRef[nodeID]
	return ref, ok
}

func (r *RefResolver) Clear() {
	r.refToNode = make(map[string]int64)
	r.nodeToRef = make(map[int64]string)
}

func (r *RefResolver) Count() int {
	return len(r.refToNode)
}

func IsRef(s string) bool {
	if !strings.HasPrefix(s, "@e") {
		return false
	}
	_, err := strconv.Atoi(s[2:])
	return err == nil
}

func ParseRef(ref string) (int, error) {
	if !IsRef(ref) {
		return 0, fmt.Errorf("invalid ref format: %s", ref)
	}
	n, err := strconv.Atoi(ref[2:])
	if err != nil {
		return 0, fmt.Errorf("invalid ref: %s", ref)
	}
	return n, nil
}

func FormatRef(n int) string {
	return fmt.Sprintf("@e%d", n)
}
