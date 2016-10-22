/*
   Copyright (C) 2016 The BlameWarrior Authors.

   This file is a part of BlameWarrior service.

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	binaryName     = os.Args[0]
	version        = "n/a"
	buildDate      = "n/a"
	builder        = "n/a"
	buildGoVersion = "n/a"

	args struct {
		version bool
	}
)

func init() {
	flag.BoolVar(&args.version, "version", false, "Print version and quit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\nOptions:\n", binaryName)
		flag.PrintDefaults()
	}
}

func printVersion() {
	fmt.Printf("%s version %s built with %s by %s on %s\n", binaryName, version, buildGoVersion, builder, buildDate)
}

func main() {
	flag.Parse()
	if args.version {
		printVersion()
		os.Exit(0)
	}

}
