package services

import (
	"context"
	"errors"
	"io"
)

// FakeConnector .
type FakeConnector struct {
	Local  io.ReadWriteCloser
	Remote io.ReadWriteCloser
}

// NewFakeConnector .
func NewFakeConnector(local, remote io.ReadWriteCloser) *FakeConnector {
	return &FakeConnector{
		Local:  local,
		Remote: remote,
	}
}

// FakeReadWriter .
type FakeReadWriter struct {
	io.ReadWriteCloser
}

var _ io.ReadWriteCloser = (*FakeReadWriter)(nil)

func (*FakeReadWriter) Read(p []byte) (int, error) {
	return 0, errors.New("cant read from conn")
}

// Connect .
func (c *FakeConnector) Connect(context.Context) (io.ReadWriteCloser, error) {
	return c.Local, nil
}
