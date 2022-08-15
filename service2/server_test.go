package main

import (
	"bytes"
	"errors"
	"io"
	"net"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleConn_OK(t *testing.T) {
	req, res := "12,43\r\n11,3\r\n\r\n ", "516\r\n33\r\n\r\n "
	a, b := net.Pipe()
	handleConn(b)

	buf := bytes.NewBuffer([]byte(req))

	_, err := buf.WriteTo(a)
	assert.NoError(t, err)

	_, err = buf.ReadFrom(a)
	assert.NoError(t, err)

	assert.Equal(t, buf.String(), res)
}

func TestHandleConn_UnmarshalErr(t *testing.T) {
	req := "12,43\r\noops,3\r\n\r\n "
	a, b := net.Pipe()
	errch := handleConn(b)

	buf := bytes.NewBuffer([]byte(req))

	_, err := buf.WriteTo(a)
	assert.NoError(t, err)

	err = <-errch

	assert.ErrorAs(t, err, &strconv.ErrSyntax)

}

func TestHandleConn_ErrRead(t *testing.T) {
	a, b := net.Pipe()
	var errch chan error
	a.Close()

	errch = handleConn(b)
	handleErr(errch)
	err := <-errch
	assert.ErrorIs(t, err, io.EOF)
}

type fakeReadWriter struct {
	io.ReadWriteCloser
}

var errCantWriteConn = errors.New("cant write to conn")
var _ io.ReadWriteCloser = (*fakeReadWriter)(nil)

func (*fakeReadWriter) Write(p []byte) (int, error) {
	return 0, errCantWriteConn
}

func TestHandleConn_ErrWrite(t *testing.T) {
	req := "12,43\r\n11,3\r\n\r\n "
	a, b := net.Pipe()

	errch := handleConn(&fakeReadWriter{b})

	buf := bytes.NewBuffer([]byte(req))

	_, err := buf.WriteTo(a)
	assert.NoError(t, err)

	err = <-errch
	assert.Equal(t, err, errCantWriteConn)
}

func TestUnmarshalMsg(t *testing.T) {
	testCases := []struct {
		name string
		req  string
		res  []*pair
		err  error
	}{
		{
			name: "empty",
			req:  "",
			res:  nil,
			err:  ErrNotCorrectFormat,
		},
		{
			name: "not integer first digit in pair",
			req:  "12,43\r\noops,3\r\n\r\n ",
			res:  nil,
			err:  strconv.ErrSyntax,
		},
		{
			name: "not integer second digit in pair",
			req:  "12,43\r\n11,oops\r\n\r\n ",
			res:  nil,
			err:  strconv.ErrSyntax,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			res, err := unmarshalMsg(tc.req)

			if err != tc.err && !errors.As(err, &tc.err) {
				t.Errorf("expecting %v %T, got %v, %T", tc.err, tc.err, err, err)
			}

			assert.Equal(t, tc.res, res)
		})
	}
}
