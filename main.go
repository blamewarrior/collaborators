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
	"log"
	"os"

	"github.com/blamewarrior/collaborators/blamewarrior"
	"github.com/blamewarrior/collaborators/blamewarrior/tokens"
	"github.com/blamewarrior/collaborators/github"
	"github.com/bmizerany/pat"
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

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Fatal("missing test database name (expected to be passed via ENV['DB_NAME'])")
	}

	opts := &blamewarrior.DatabaseOptions{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
	}

	db, err := blamewarrior.ConnectDatabase(dbName, opts)
	if err != nil {
		log.Fatalf("failed to establish connection with test db %s using connection string %s: %s", dbName, opts.ConnectionString(), err)
	}

	mux := pat.New()

	tokenClient := tokens.NewTokenClient("https://blamewarrior.com")
	githubClient := github.NewClient(tokenClient)

	collaboration := blamewarrior.NewCollaborationService()

	mux.Get("/:username/:repo/collaborators/fetch", NewFetchCollaboratorsHandler("blamewarrior.com", db, collaboration,
		githubClient))
	mux.Post("/:username/:repo/collaborators", NewAddCollaboratorHandler("blamewarrior.com", db, collaboration))
	mux.Get("/:username/:repo/collaborators", NewListCollaboratorHandler("blamewarrior.com", collaboration))
	mux.Put("/:username/:repo/collaborators", NewEditCollaboratorHandler("blamewarrior.com", collaboration))
	mux.Del("/:username/:repo/collaborators/:collaborator", NewDisconnectCollaboratorHandler("blamewarrior.com", db, collaboration))
}
