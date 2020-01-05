// This file is part of gosyncmaildir.
//
// Copyright (C) 2020 Julian Andres Klode <jak@jak-linux.org>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"os"

	"github.com/julian-klode/gosyncmaildir/gsmd"
)

func main() {
	a := gsmd.BuildTree(os.Args[1])
	b := gsmd.BuildTree(os.Args[2])
	base := gsmd.BuildTree(os.Args[3])

	res := gsmd.Merge(a, b, base, gsmd.MergeOpts{})

	d := gsmd.DiffTree(b, res)

	for _, node := range d.Add {
		fmt.Println("Add", node)
		if err := os.Symlink(os.Args[1]+"/"+node.Dir+"/"+node.ID+":"+node.Flags, os.Args[2]+"/"+node.Dir+"/"+node.ID+":"+node.Flags); err != nil {
			panic(err)
		}
	}

	for _, node := range d.Del {
		fmt.Println("Del", node)
	}

	for _, node := range d.Mod {
		old := b.Nodes[node.ID]
		if old.Dir != node.Dir || old.Flags != node.Flags {
			if old.Dir != node.Dir {
				fmt.Println("Dir", node.ID, old.Dir, node.Dir)
			}
			if old.Flags != node.Flags {
				fmt.Println("Flags", node.ID, old.Flags, node.Flags)
			}
			if err := os.Rename(os.Args[2]+"/"+old.Dir+"/"+node.ID+":"+old.Flags, os.Args[2]+"/"+node.Dir+"/"+node.ID+":"+node.Flags); err != nil {
				panic(err)
			}
		}
		if old.ModTime != node.ModTime {
			fmt.Println("Time", node.ID, old.ModTime, node.ModTime)
			if err := os.Chtimes(os.Args[2]+"/"+node.Dir+"/"+node.ID+":"+node.Flags, node.ModTime, node.ModTime); err != nil {
				panic(err)
			}
		}

	}

}
