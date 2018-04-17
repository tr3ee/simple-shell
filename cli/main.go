package main

import (
	"flag"
	"io"
	"net"
	"os"

	"github.com/tr3ee/go-link"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var (
	network, remote string
)

func init() {
	flag.StringVar(&network, "net", "tcp", "Specifies the network type for connection.")
	flag.StringVar(&remote, "addr", "localhost:6666", "Specify the address and port of the remote connection.")
	flag.Parse()
}

func main() {
	conn, err := net.Dial(network, remote)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	stdio := NewMergedIO(os.Stdin, os.Stdout)
	link.TwoWayLink(nil, conn, stdio, nil, nil)
}

type MergedIO struct {
	r  io.Reader
	w  io.Writer
	rc io.ReadCloser
	wc io.WriteCloser
}

func NewMergedIO(rc io.ReadCloser, wc io.WriteCloser) *MergedIO {
	r := transform.NewReader(rc, simplifiedchinese.GBK.NewEncoder())
	w := transform.NewWriter(wc, simplifiedchinese.GBK.NewDecoder())
	return &MergedIO{r, w, rc, wc}
}

func (io *MergedIO) Read(p []byte) (int, error) {
	return io.r.Read(p)
}

func (io *MergedIO) Write(p []byte) (int, error) {
	return io.w.Write(p)
}

func (io *MergedIO) Close() error {
	err1 := io.rc.Close()
	err2 := io.wc.Close()
	if err1 != nil {
		return err1
	}
	return err2
}
