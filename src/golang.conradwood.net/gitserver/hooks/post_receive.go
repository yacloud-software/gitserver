package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	GITSERVER_TCP_PORT = 5023
)

type PostReceive struct {
	mydir   string
	entries []*gitEntry
	ev      *Environment
}

type gitEntry struct {
	oldrev string
	newrev string
	ref    string
}

func (p *PostReceive) Process(e *Environment) error {
	p.ev = e
	var err error
	p.mydir, err = os.Getwd()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		err := scanner.Err()
		if err != nil {
			fmt.Printf("Failed to read from stdin: %s", err)
		}
		line := scanner.Text()
		fmt.Printf("Scanned: \"%s\"\n", line)
		fs := strings.Fields(line)
		if len(fs) != 3 {
			return fmt.Errorf("Failed to scan line from git \"%s\"", line)
		}
		ge := &gitEntry{oldrev: fs[0], newrev: fs[1], ref: fs[2]}
		p.entries = append(p.entries, ge)
	}

	// this really should be a streaming grpc service
	tcp_port, err := strconv.Atoi(os.Getenv("GITSERVER_TCP_PORT"))
	if err != nil {
		fmt.Printf("tcpport in environment invalid (using default): %s\n", err)
	}
	if tcp_port == 0 {
		tcp_port = GITSERVER_TCP_PORT
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", tcp_port))
	if err != nil {
		return err
	}
	defer conn.Close()

	//	send the stuff to server
	for _, ge := range p.entries {
		line := fmt.Sprintf("%s %s %s %s %s %s\n", p.mydir, ge.ref, ge.oldrev, ge.newrev, ge.ref, os.Getenv("GITINFO"))
		_, err = conn.Write([]byte(line))
		if err != nil {
			return err
		}
	}
	buf := make([]byte, 8192)
	for {
		n, err := conn.Read(buf)
		if n != 0 {
			pbuf := buf[:n]
			fmt.Printf("%s", string(pbuf))
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

