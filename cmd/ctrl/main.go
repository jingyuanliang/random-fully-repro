/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package main is the entrypoint for ctrl.
package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/jingyuanliang/random-fully-repro/pkg/version"
)

var (
	startblock string

	file *os.File
)

func init() {
	flag.StringVar(&startblock, "startblock", "/tmp/startblock/lock", "startblock lock")
	flag.Parse()
}

func block() {
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		log.Fatalf("lock startblock %q: %v\n", startblock, err)
	}

	log.Printf("startblock %q is now locked; waiting for stdin before unlocking\n", startblock)
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_UN); err != nil {
		log.Fatalf("unlock startblock %q: %v\n", startblock, err)
	}

	log.Printf("startblock %q will remain unlocked for 1 second\n", startblock)
	time.Sleep(time.Second)
}

func main() {
	log.Printf("version: %s\n", version.Version)

	var err error
	file, err = os.OpenFile(startblock, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("open startblock: %v\n", err)
	}

	for {
		block()
	}
}
