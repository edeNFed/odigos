package api

import (
	"github.com/odigos-io/odigos/profiles-viewer/internal/models"
)

// BuildFlameGraph converts an aggregated profile into a flame graph tree structure.
func BuildFlameGraph(profile *models.AggregatedProfile) *models.FlameGraphNode {
	root := &models.FlameGraphNode{
		Name:     "root",
		Value:    0,
		Children: make([]*models.FlameGraphNode, 0),
	}

	if profile == nil {
		return root
	}

	// Process each stack
	for _, stack := range profile.Stacks {
		addStackToTree(root, stack, profile.Functions)
	}

	// Calculate root value as sum of all children
	for _, child := range root.Children {
		root.Value += child.Value
	}

	return root
}

// addStackToTree adds a stack trace to the flame graph tree.
func addStackToTree(root *models.FlameGraphNode, stack *models.StackInfo, functions map[int64]*models.FunctionInfo) {
	if len(stack.FunctionIDs) == 0 {
		return
	}

	current := root

	// Process stack from bottom (root) to top (leaf)
	// Stack is stored leaf-to-root, so we iterate in reverse
	for i := len(stack.FunctionIDs) - 1; i >= 0; i-- {
		funcID := stack.FunctionIDs[i]
		fn, ok := functions[funcID]
		if !ok {
			continue
		}

		name := formatFunctionName(fn)

		// Find or create child
		var child *models.FlameGraphNode
		for _, c := range current.Children {
			if c.Name == name {
				child = c
				break
			}
		}

		if child == nil {
			child = &models.FlameGraphNode{
				Name:     name,
				Value:    0,
				Children: make([]*models.FlameGraphNode, 0),
			}
			current.Children = append(current.Children, child)
		}

		child.Value += stack.SampleCount
		current = child
	}
}

// formatFunctionName creates a display name for a function.
func formatFunctionName(fn *models.FunctionInfo) string {
	if fn.Filename != "" && fn.Line > 0 {
		return fn.Name + " (" + fn.Filename + ":" + formatLine(fn.Line) + ")"
	}
	if fn.Filename != "" {
		return fn.Name + " (" + fn.Filename + ")"
	}
	return fn.Name
}

// formatLine converts a line number to string.
func formatLine(line int64) string {
	if line <= 0 {
		return "?"
	}
	// Simple int to string without importing strconv
	if line == 0 {
		return "0"
	}
	var digits []byte
	for line > 0 {
		digits = append([]byte{byte('0' + line%10)}, digits...)
		line /= 10
	}
	return string(digits)
}

// SortFlameGraph sorts flame graph children by value (descending).
func SortFlameGraph(node *models.FlameGraphNode) {
	if node == nil || len(node.Children) == 0 {
		return
	}

	// Simple bubble sort for children (usually not many)
	for i := 0; i < len(node.Children); i++ {
		for j := i + 1; j < len(node.Children); j++ {
			if node.Children[j].Value > node.Children[i].Value {
				node.Children[i], node.Children[j] = node.Children[j], node.Children[i]
			}
		}
	}

	// Recursively sort children
	for _, child := range node.Children {
		SortFlameGraph(child)
	}
}
