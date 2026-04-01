package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// PrintTree prints a tree view of files that will be generated.
func PrintTree(outputDir string, files []string) {
	if len(files) == 0 {
		return
	}

	root := filepath.Base(outputDir)
	fmt.Println(root + "/")

	// Sort files for deterministic output.
	sorted := make([]string, len(files))
	copy(sorted, files)
	sort.Strings(sorted)

	// Build a tree structure.
	type node struct {
		name     string
		children map[string]*node
		order    []string
	}

	newNode := func(name string) *node {
		return &node{name: name, children: make(map[string]*node)}
	}

	rootNode := newNode(root)
	for _, f := range sorted {
		parts := strings.Split(filepath.ToSlash(f), "/")
		current := rootNode
		for _, p := range parts {
			if _, ok := current.children[p]; !ok {
				current.children[p] = newNode(p)
				current.order = append(current.order, p)
			}
			current = current.children[p]
		}
	}

	// Print the tree recursively.
	var printNode func(n *node, prefix string)
	printNode = func(n *node, prefix string) {
		for i, name := range n.order {
			child := n.children[name]
			isLast := i == len(n.order)-1

			connector := "├── "
			if isLast {
				connector = "└── "
			}

			suffix := ""
			if len(child.children) > 0 {
				suffix = "/"
			}

			fmt.Println(prefix + connector + name + suffix)

			nextPrefix := prefix + "│   "
			if isLast {
				nextPrefix = prefix + "    "
			}
			printNode(child, nextPrefix)
		}
	}

	printNode(rootNode, "")
}
