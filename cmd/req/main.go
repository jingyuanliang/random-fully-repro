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

// Package main is the entrypoint for req.
package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/jingyuanliang/random-fully-repro/pkg/version"
)

var (
	client     string
	server     string
	request    string
	startblock string

	c, s *net.TCPAddr
)

func init() {
	flag.StringVar(&client, "client", ":33333", "client addr")
	flag.StringVar(&server, "server", "34.160.111.145:80", "server addr")
	flag.StringVar(&request, "request", "GET /all HTTP/1.0\r\nHost: ifconfig.me\r\n\r\n", "request msg")
	flag.StringVar(&startblock, "startblock", "/tmp/startblock/lock", "startblock lock")
	flag.Parse()
}

func wait() {
	file, err := os.OpenFile(startblock, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("open startblock: %v\n", err)
	}

	for {
		err := syscall.Flock(int(file.Fd()), syscall.LOCK_SH|syscall.LOCK_NB)
		if errors.Is(err, syscall.EAGAIN) {
			log.Printf("waiting for startblock %q to be unlocked: %v\n", startblock, err)
			break
		}
		if err != nil {
			log.Fatalf("failed to probe startblock %q: %v\n", startblock, err)
		}
		log.Printf("startblock %q is not yet locked\n", startblock)
		if err := syscall.Flock(int(file.Fd()), syscall.LOCK_UN); err != nil {
			log.Fatalf("unlock startblock %q: %v\n", startblock, err)
		}
		time.Sleep(time.Second)
	}

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_SH); err != nil {
		log.Fatalf("lock startblock %q: %v\n", startblock, err)
	}

	log.Printf("startblock %q is now unlocked\n", startblock)

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_UN); err != nil {
		log.Fatalf("unlock startblock %q: %v\n", startblock, err)
	}
}

func req() error {
	conn, err := net.DialTCP("tcp", c, s)
	if err != nil {
		log.Printf("dial error: %v\n", err)
		return err
	}
	defer conn.Close()
	log.Printf("connected %s -> %s\n", conn.LocalAddr(), conn.RemoteAddr())

	_, err = conn.Write([]byte(request))
	if err != nil {
		log.Printf("write error: %v\n", err)
		return err
	}

	resp, err := io.ReadAll(conn)
	if err != nil {
		log.Printf("read error: %v\n", err)
		return err
	}

	log.Printf("got: %q\n", resp)
	return nil
}

func main() {
	log.Printf("version: %s\n", version.Version)

	var err error

	c, err = net.ResolveTCPAddr("tcp", client)
	if err != nil {
		log.Fatalf("resolve client %q: %v\n", client, err)
	}

	s, err = net.ResolveTCPAddr("tcp", server)
	if err != nil {
		log.Fatalf("resolve server %q: %v\n", server, err)
	}

	for {
		wait()

		i := 0
		for ; req() != nil; i++ {}
		log.Printf("request completed with %d retries\n", i)
	}
}
