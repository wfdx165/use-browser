package snapshot

import (
	"encoding/json"
	"fmt"
	"strings"
)

type FormatOptions struct {
	InteractiveOnly bool
	URLs            bool
	Compact         bool
	Depth           int
	Selector        string
}

func FormatText(snapshot *Snapshot, opts FormatOptions) string {
	if snapshot.Root == nil {
		return "(empty)"
	}

	var sb strings.Builder
	formatNode(&sb, snapshot.Root, snapshot, opts, "", true, 0, opts.Depth)
	return sb.String()
}

func formatNode(sb *strings.Builder, node *Node, snapshot *Snapshot, opts FormatOptions, prefix string, isLast bool, currentDepth, maxDepth int) {
	if maxDepth > 0 && currentDepth > maxDepth {
		return
	}

	if currentDepth > 0 {
		if isLast {
			sb.WriteString(prefix)
			sb.WriteString("└─ ")
		} else {
			sb.WriteString(prefix)
			sb.WriteString("├─ ")
		}
	}

	sb.WriteString(fmt.Sprintf("[%s] %s", strings.TrimPrefix(node.Ref, "@"), node.Role))

	if node.Name != "" {
		sb.WriteString(fmt.Sprintf(" %q", node.Name))
	}

	if node.Value != "" {
		sb.WriteString(fmt.Sprintf(" value=%q", node.Value))
	}

	sb.WriteString("\n")

	childPrefix := prefix
	if currentDepth > 0 {
		if isLast {
			childPrefix += "   "
		} else {
			childPrefix += "│  "
		}
	}

	for i, childRef := range node.Children {
		child, ok := snapshot.Nodes[childRef]
		if !ok {
			continue
		}
		isLastChild := i == len(node.Children)-1
		formatNode(sb, child, snapshot, opts, childPrefix, isLastChild, currentDepth+1, maxDepth)
	}
}

func FormatJSON(snapshot *Snapshot, opts FormatOptions) ([]byte, error) {
	refs := make(map[string]map[string]string, len(snapshot.Nodes))
	for ref, node := range snapshot.Nodes {
		refs[ref] = map[string]string{
			"role": node.Role,
			"name": node.Name,
		}
	}

	tree := FormatText(snapshot, opts)

	output := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"snapshot": tree,
			"refs":     refs,
		},
	}

	return json.Marshal(output)
}
