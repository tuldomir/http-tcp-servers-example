package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

// Server .
type Server struct {
	listener net.Listener
}

// New .
func New(host, port string) (*Server, error) {
	listener, err := net.Listen("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return nil, fmt.Errorf("cant create listener %w", err)
	}

	return &Server{listener: listener}, nil
}

// Stop .
func (s *Server) Stop() error {
	return s.listener.Close()
}

// Run .
func (s *Server) Run() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("new conn")
		conn.SetReadDeadline(time.Now().Add(1 * time.Minute))

		errCh := handleConn(conn)
		handleErr(errCh)
	}
}

func handleErr(ch chan error) {
	go func() {
		fmt.Println(<-ch)
	}()
}

func handleConn(conn io.ReadWriteCloser) chan error {
	errch := make(chan error)

	go func() {
		defer close(errch)
		defer conn.Close()

		buf := bufio.NewReader(conn)
		bs, err := buf.ReadBytes(msgdelim)

		if err != nil {
			errch <- err
			time.Sleep(1 * time.Second)
			return
		}

		pairs, err := unmarshalMsg(string(bs))
		if err != nil {
			errch <- err
			return
		}

		muls := mulPairs(pairs)
		out := marshalMsg(muls)

		if _, err = conn.Write([]byte(out)); err != nil {
			errch <- err
		}
	}()

	return errch
}

type pair struct {
	a, b int
}

func unmarshalMsg(str string) ([]*pair, error) {
	pairs := make([]*pair, 0)

	str = strings.TrimSuffix(str, pairsep+eof)

	strs := strings.Split(str, "\r\n")

	for _, v := range strs {
		strpair := strings.Split(v, ",")

		if len(strpair) != 2 {
			return nil, ErrNotCorrectFormat
		}

		a, err := strconv.Atoi(strpair[0])
		if err != nil {
			return nil, fmt.Errorf("not correct format %w", err)
		}

		b, err := strconv.Atoi(strpair[1])
		if err != nil {
			return nil, fmt.Errorf("not correct format %w", err)
		}

		pairs = append(pairs, &pair{a: a, b: b})
	}

	return pairs, nil
}

func mulPairs(pairs []*pair) []int {
	res := make([]int, len(pairs))

	for i, v := range pairs {
		res[i] = v.a * v.b
	}

	return res
}

func marshalMsg(muls []int) string {

	var builder strings.Builder

	for _, v := range muls {
		str := strconv.Itoa(v)

		builder.WriteString(str)
		builder.WriteString("\r\n")
	}

	builder.WriteString("\r\n ")

	return builder.String()
}
