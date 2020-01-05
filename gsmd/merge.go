// This file is part of gosyncmaildir.
//
// Copyright (C) 2020 Julian Andres Klode <jak@jak-linux.org>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package gsmd

// MergeOpts contains options for the merge algorithm.
type MergeOpts struct {
}

// Merge produces a new merged tree, incorporating the changes from base to A,
// and base to B; resolving any conflicts by preferring A.
//
// Conflict resolution:
// - A adds, B {adds, deletes} => A wins
// - A adds, B modifies => B wins
// - A deletes, B {adds, modifies} => B wins
// - A modifies, B {adds, modifies, deletes} => A wins
func Merge(a, b, base Tree, opts MergeOpts) Tree {
	diffA := DiffTree(base, a)
	diffB := DiffTree(base, b)

	res := base.Copy()
	for _, del := range diffB.Del {
		//fmt.Println("Deleting", del.ID)
		delete(res.Nodes, del.ID)
	}

	for _, del := range diffA.Del {
		//fmt.Println("Deleting", del.ID)
		delete(res.Nodes, del.ID)
	}

	for _, add := range diffB.Add {
		//fmt.Println("Adding A", add.ID)
		res.Nodes[add.ID] = add
	}
	for _, add := range diffA.Add {
		//fmt.Println("Adding B", add.ID)
		res.Nodes[add.ID] = add
	}

	for _, mod := range diffB.Mod {
		//fmt.Println("Modifying B", mod.ID, res.Nodes[mod.ID], mod)
		res.Nodes[mod.ID] = mod
	}
	for _, mod := range diffA.Mod {
		//fmt.Println("Modifying A", mod.ID, res.Nodes[mod.ID], mod)
		res.Nodes[mod.ID] = mod
	}

	return res
}
