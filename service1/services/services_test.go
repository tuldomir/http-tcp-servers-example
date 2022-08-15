package services

import (
	"bufio"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"net"
	"service1/database"
	"service1/models"
	"strconv"
	"testing"

	"github.com/go-redis/redis/v8"

	"github.com/alicebob/miniredis"
	"github.com/stretchr/testify/assert"
)

func TestServiceMulStringVal_RemoteUnavail(t *testing.T) {
	remoteHost, remotePort := "localhost", "8888"
	testCase := struct {
		pairs []*models.Pair
		res   map[string]int
		err   error
	}{
		pairs: []*models.Pair{{A: "12", B: "43", Key: "x"}},
		res:   nil,
		err:   &net.OpError{},
	}

	serv := NewTService(nil,
		NewTCPConnector(net.JoinHostPort(remoteHost, remotePort)))

	res, err := serv.MulStringVal(context.Background(), testCase.pairs)
	assert.ErrorAs(t, err, &testCase.err)
	assert.Equal(t, testCase.res, res)
}

func TestServiceMulStringVal_RemoteWriteErr(t *testing.T) {
	testCase := struct {
		pairs []*models.Pair
		res   map[string]int
		err   error
	}{
		pairs: []*models.Pair{{A: "12", B: "43", Key: "x"}},
		res:   nil,
		err:   &net.OpError{},
	}

	fakeconn := NewFakeConnector(net.Pipe())
	fakeconn.Remote.Close()

	srv := NewTService(nil, fakeconn)
	m, err := srv.MulStringVal(context.Background(), testCase.pairs)

	assert.ErrorAs(t, err, &testCase.err)
	assert.Equal(t, testCase.res, m)
}

func TestServiceMulStringVal_RemoteReadErr(t *testing.T) {
	testCase := struct {
		pairs []*models.Pair
		res   map[string]int
		err   error
	}{
		pairs: []*models.Pair{{A: "12", B: "43", Key: "x"}},
		res:   nil,
		err:   &net.OpError{},
	}

	a, b := net.Pipe()
	go func() {
		_, err := ioutil.ReadAll(b)
		if err != nil {
			t.Error()
		}
	}()

	fakeconn := NewFakeConnector(&FakeReadWriter{a}, b)

	srv := NewTService(nil, fakeconn)
	m, err := srv.MulStringVal(context.Background(), testCase.pairs)

	assert.ErrorAs(t, err, &testCase.err)
	assert.Equal(t, testCase.res, m)
}

func TestUnmarshalMsg(t *testing.T) {
	testCases := []struct {
		name string
		keys []string
		str  string
		res  map[string]int
		err  error
	}{
		{
			name: "ok",
			keys: []string{"x", "y"},
			str:  "516\r\n33\r\n\r\n ",
			res:  map[string]int{"x": 516, "y": 33},
			err:  nil,
		},

		{
			name: "zero message",
			keys: []string{"x", "y"},
			str:  "",
			res:  nil,
			err:  ErrNotCorrectFormat,
		},
		{
			name: "invalid syntax",
			keys: []string{"x", "y"},
			str:  "516\r\noops\r\n\r\n ",
			res:  nil,
			err:  strconv.ErrSyntax,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			m, err := UnmarshalMsg(tc.keys, tc.str)

			if err != tc.err && !errors.As(err, &tc.err) {
				t.Fatalf("expecting err %v, %T, got %v, %T", err, err, tc.err, tc.err)
			}

			assert.Equal(t, tc.res, m)
		})
	}
}

func TestMulStringVal_UnMarshalErr(t *testing.T) {
	testCase := struct {
		pairs     []*models.Pair
		remoteRes string
		res       map[string]int
	}{
		pairs:     []*models.Pair{{A: "12", B: "43", Key: "x"}},
		remoteRes: "516\r\noops\r\n\r\n ",
		res:       nil,
	}

	fake := NewFakeConnector(net.Pipe())
	serv := NewTService(nil, fake)

	go func() {
		reader := bufio.NewReader(fake.Remote)
		_, err := reader.ReadBytes(' ')
		assert.NoError(t, err)

		writer := bufio.NewWriter(fake.Remote)
		_, err = writer.WriteString(testCase.remoteRes)

		err = writer.Flush()
		assert.NoError(t, err)
		fake.Remote.Close()
	}()

	res, err := serv.MulStringVal(context.Background(), testCase.pairs)
	assert.Error(t, err)
	assert.Equal(t, testCase.res, res)
}

func TestHashString(t *testing.T) {

	testCases := []struct {
		name string
		str  string
		key  string
		res  func() string
	}{
		{
			name: "ok",
			str:  "test",
			key:  "test123",
			res: func() string {
				hm := hmac.New(sha512.New, []byte("test123"))
				hm.Write([]byte("test"))
				return hex.EncodeToString(hm.Sum(nil))
			},
		},
	}

	serv := NewTService(nil, nil)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := serv.HashString(context.Background(), tc.str, tc.key)

			assert.Equal(t, tc.res(), res)
		})
	}
}

func TestIncrementBy(t *testing.T) {
	testCases := []struct {
		name string
		key  string
		val  int64
		res  map[string]int64
		err  error
	}{
		{
			name: "ok",
			key:  "test",
			val:  12,
			res:  map[string]int64{"test": 12},
			err:  nil,
		},
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisServer.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisServer.Addr(),
	})

	assert.NoError(t, err)

	db := database.NewDB(redisClient)
	defer db.Stop()

	srv := NewTService(db, nil)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			m, err := srv.IncrementBy(context.Background(), tc.key, tc.val)
			assert.Equal(t, tc.res, m)

			assert.NoError(t, err)
		})
	}
}
