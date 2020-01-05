// This file is part of gosyncmaildir.
//
// Copyright (C) 2020 Julian Andres Klode <jak@jak-linux.org>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package gsmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Tree is a tree
type Tree struct {
	Nodes map[string]Node
}

// Node is a node
type Node struct {
	Dir     string
	ID      string
	ModTime time.Time
	Flags   string
}

// TreeDifference stores the difference between two trees
type TreeDifference struct {
	// Add are added nodes
	Add []Node
	// Del are nodes that have been deleted
	Del []Node
	// Mod are nodes that have been moved, or flags changed
	Mod []Node
}

// Copy copies a tree
func (tree *Tree) Copy() Tree {
	var res Tree

	res.Nodes = make(map[string]Node)
	for id, node := range tree.Nodes {
		res.Nodes[id] = node
	}

	return res
}

// BuildTree builds a tree
func BuildTree(root string) Tree {
	var tree Tree

	tree.Nodes = make(map[string]Node)

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !strings.Contains(path, ":2") {
			return nil
		}
		dir, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}

		id := strings.Split(filepath.Base(path), ":")[0]
		node := Node{
			Dir:     dir,
			ID:      id,
			ModTime: info.ModTime(),
			Flags:   strings.Split(filepath.Base(path), ":")[1],
		}
		if ex, ok := tree.Nodes[id]; ok {
			fmt.Fprintf(os.Stderr, "%s is duplicate of %s\n", node, ex)
		}
		tree.Nodes[id] = node

		return nil

	})

	return tree
}

// DiffTree calculates the difference between two trees
func DiffTree(a, b Tree) TreeDifference {
	var diff TreeDifference
	for id, node := range b.Nodes {
		nodea, inA := a.Nodes[id]
		if !inA {
			diff.Add = append(diff.Add, node)
		} else if nodea != node {
			diff.Mod = append(diff.Mod, node)
		}
	}

	for id, node := range a.Nodes {
		if _, inB := b.Nodes[id]; !inB {
			diff.Del = append(diff.Del, node)
		}
	}

	return diff
}
