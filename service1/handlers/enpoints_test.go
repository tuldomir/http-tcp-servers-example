package handlers

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"service1/database"
	"service1/services"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"

	"github.com/stretchr/testify/assert"
)

func TestHandlerHashStringHandler(t *testing.T) {

	handler := NewHandler(services.NewTService(nil, nil))

	testCases := []struct {
		name         string
		req          string
		res          func() string
		expectedCode int
	}{
		{
			name: "ok",
			req: `{
					"s": "test",
					"key": "test123"
				 }`,
			res: func() string {
				hm := hmac.New(sha512.New, []byte("test123"))
				hm.Write([]byte("test"))
				return hex.EncodeToString(hm.Sum(nil))
			},
			expectedCode: http.StatusOK,
		},

		{
			name: "incorrect msg format",
			req: `{
					"something": "test",
					"key": "test123"
				 }`,
			res: func() string {
				msg, _ := json.Marshal(map[string]string{"error": ErrNotCorrectMsg.Error()})
				return string(msg) + "\n"
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "incorrect msg data",
			req: `{
					"s": 1,
					"key": "test123"
				 }`,
			res: func() string {
				msg, _ := json.Marshal(map[string]string{"error": ErrNotCorrectMsg.Error()})
				return string(msg) + "\n"
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			b := &bytes.Buffer{}
			b.WriteString(tc.req)

			req, _ := http.NewRequest(http.MethodPost, "/test2", b)
			req.Header.Set("Content-Type", "application/json")

			handler.HashStringHandler().ServeHTTP(rec, req)
			result := rec.Result()

			assert.Equal(t, tc.expectedCode, result.StatusCode)

			bs := rec.Body.Bytes()
			res := tc.res()

			assert.Equal(t, res, string(bs))

		})
	}
}

func TestHandlerMulStringValHandler_Ok(t *testing.T) {
	testcase := struct {
		req          string
		res          string
		remoteReq    string
		remoteRes    string
		expectedCode int
	}{
		req: `[
			{
			"a": "12",
			"b": "43",
			"key": "x"
			},
			{
			"a": "11",
			"b": "3",
			"key": "y"
			}
			]`,
		res:          `{"x":516,"y":33}` + "\n",
		remoteReq:    "12,43\r\n11,3\r\n\r\n ",
		remoteRes:    "516\r\n33\r\n\r\n ",
		expectedCode: http.StatusOK,
	}

	fake := services.NewFakeConnector(net.Pipe())
	handler := NewHandler(services.NewTService(nil, fake))

	go func() {
		reader := bufio.NewReader(fake.Remote)
		_, err := reader.ReadBytes(' ')
		assert.NoError(t, err)

		writer := bufio.NewWriter(fake.Remote)
		_, err = writer.WriteString("516\r\n33\r\n\r\n ")

		err = writer.Flush()
		assert.NoError(t, err)
		fake.Remote.Close()
	}()

	rec := httptest.NewRecorder()

	b := &bytes.Buffer{}
	_, err := b.WriteString(testcase.req)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodPost, "/test3", b)
	req.Header.Set("Content-Type", "application/json")

	handler.MulStringValHandler().ServeHTTP(rec, req)
	result := rec.Result()

	assert.Equal(t, testcase.expectedCode, result.StatusCode)

	bs := rec.Body.Bytes()

	assert.Equal(t, testcase.res, string(bs))

}

// THIS ONE ALSO CAN BE USED instead of previous one

// func TestHandlerTest3_Ok(t *testing.T) {
// 	testcase := struct {
// 		remoteHost   string
// 		remotePort   string
// 		req          string
// 		res          string
// 		remoteReq    string
// 		remoteRes    string
// 		expectedCode int
// 	}{
// 		remoteHost: "localhost",
// 		remotePort: "5555",
// 		req: `[
// 			{
// 			"a": "12",
// 			"b": "43",
// 			"key": "x"
// 			},
// 			{
// 			"a": "11",
// 			"b": "3",
// 			"key": "y"
// 			}
// 			]`,
// 		res:          `{"x":516,"y":33}` + "\n",
// 		remoteReq:    "12,43\r\n11,3\r\n\r\n ",
// 		remoteRes:    "516\r\n33\r\n\r\n ",
// 		expectedCode: http.StatusOK,
// 	}

// 	handler := NewHandler(NewTService(nil, NewTCPConnector(net.JoinHostPort(testcase.remoteHost, testcase.remotePort))))

// 	remote := func() {
// 		ln, err := net.Listen("tcp", net.JoinHostPort(testcase.remoteHost, testcase.remotePort))
// 		defer ln.Close()
// 		assert.NoError(t, err, "cant crearte listener")

// 		for {
// 			conn, err := ln.Accept()
// 			assert.NoError(t, err, "cant accept connection")

// 			go func(conn net.Conn) {
// 				defer conn.Close()
// 				buf := bufio.NewReader(conn)
// 				bs, err := buf.ReadBytes(' ')

// 				assert.NoError(t, err, "cant read from connection")
// 				assert.Equal(t, string(bs), testcase.remoteReq)

// 				_, err = conn.Write([]byte(testcase.remoteRes))
// 				assert.NoError(t, err, "cant write to connection")

// 			}(conn)
// 		}
// 	}

// 	go remote()

// 	rec := httptest.NewRecorder()

// 	b := &bytes.Buffer{}
// 	b.WriteString(testcase.req)

// 	req, _ := http.NewRequest(http.MethodPost, "/test3", b)
// 	req.Header.Set("Content-Type", "application/json")

// 	handler.MulStringValHandler().ServeHTTP(rec, req)
// 	result := rec.Result()

// 	assert.Equal(t, testcase.expectedCode, result.StatusCode)

// 	bs := rec.Body.Bytes()

// 	assert.Equal(t, testcase.res, string(bs))

// }

func TestHandlerMulStringValHandler_IncorrectRequestMsg(t *testing.T) {
	testCases := []struct {
		name         string
		req          string
		res          func() string
		expectedCode int
	}{
		{
			name: "invalid field values",
			req: `[
			{
			"a": 12,
			"b": 43,
			"key": "x"
			}
			]`,
			res: func() string {
				msg, _ := json.Marshal(map[string]string{"error": ErrNotCorrectMsg.Error()})
				return string(msg) + "\n"
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "empty field",
			req: `[
			{
			"b": "43",
			"key": "x"
			}
			]`,
			res: func() string {
				msg, _ := json.Marshal(map[string]string{"error": ErrNotCorrectMsg.Error()})
				return string(msg) + "\n"
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	handler := NewHandler(services.NewTService(nil, nil))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			b := &bytes.Buffer{}
			b.WriteString(tc.req)

			req, _ := http.NewRequest(http.MethodPost, "/test3", b)
			req.Header.Set("Content-Type", "application/json")

			handler.MulStringValHandler().ServeHTTP(rec, req)
			result := rec.Result()

			assert.Equal(t, tc.expectedCode, result.StatusCode)

			bs := rec.Body.Bytes()
			res := tc.res()

			assert.Equal(t, res, string(bs))

		})
	}
}

func TestHandlerMulStringValHandler_RemoteErr(t *testing.T) {
	remoteHost, remotePort := "localhost", "8888"
	testCase := struct {
		req          string
		expectedCode int
	}{

		req: `[
			{
			"a": "12",
			"b": "43",
			"key": "x"
			}
			]`,
		expectedCode: http.StatusInternalServerError,
	}

	handler := NewHandler(services.NewTService(nil,
		services.NewTCPConnector(net.JoinHostPort(remoteHost, remotePort))))

	rec := httptest.NewRecorder()

	b := &bytes.Buffer{}
	b.WriteString(testCase.req)

	req, _ := http.NewRequest(http.MethodPost, "/test3", b)
	req.Header.Set("Content-Type", "application/json")

	handler.MulStringValHandler().ServeHTTP(rec, req)
	result := rec.Result()

	assert.Equal(t, testCase.expectedCode, result.StatusCode)
}

func TestHandlerIncrementByHandler(t *testing.T) {
	testCases := []struct {
		name         string
		req          string
		res          func() string
		expectedCode int
	}{
		{
			name: "ok",
			req:  `{"key": "test","val": 12}`,
			res: func() string {
				return `{"test":12}` + "\n"
			},
			expectedCode: http.StatusOK,
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

	handler := NewHandler(services.NewTService(db, nil))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			b := &bytes.Buffer{}
			b.WriteString(tc.req)

			req, _ := http.NewRequest(http.MethodPost, "/test1", b)
			req.Header.Set("Content-Type", "application/json")

			handler.IncrementByHandler().ServeHTTP(rec, req)

			result := rec.Result()
			assert.Equal(t, tc.expectedCode, result.StatusCode)

			bs := rec.Body.Bytes()
			res := tc.res()

			assert.Equal(t, res, string(bs))
		})
	}
}

func TestHandlerIncrementByHandler_IncorrectReqMsg(t *testing.T) {
	testCases := []struct {
		name         string
		req          string
		expectedCode int
	}{
		{
			name:         "invalid msg structure",
			req:          `{"key": "test","something": 12}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid msg field type",
			req:          `{"key": "test","val": "12"}`,
			expectedCode: http.StatusInternalServerError,
		},
	}

	handler := NewHandler(services.NewTService(nil, nil))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			b := &bytes.Buffer{}
			b.WriteString(tc.req)

			req, _ := http.NewRequest(http.MethodPost, "/test1", b)
			req.Header.Set("Content-Type", "application/json")

			handler.IncrementByHandler().ServeHTTP(rec, req)

			result := rec.Result()
			assert.Equal(t, tc.expectedCode, result.StatusCode)
		})
	}
}

func TestHandlerIncrementByHandler_DBError(t *testing.T) {
	msg := `{"key": "test","val": 12}`
	expectedCode := http.StatusInternalServerError

	handler := NewHandler(services.NewTService(&database.FakeDB{}, nil))
	rec := httptest.NewRecorder()

	b := &bytes.Buffer{}
	b.WriteString(msg)

	req, _ := http.NewRequest(http.MethodPost, "/test1", b)
	req.Header.Set("Content-Type", "application/json")

	handler.IncrementByHandler().ServeHTTP(rec, req)

	result := rec.Result()
	assert.Equal(t, expectedCode, result.StatusCode)
}
