package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"service1/database"
	"service1/models"
	"strconv"
	"strings"
)

const (
	pairsep  = "\r\n"
	digitsep = ","
	eof      = ' '
)

// ErrNotCorrectFormat .
var ErrNotCorrectFormat = errors.New("not correct format")

// Service .
type Service interface {
	IncrementBy(context.Context, string, int64) (map[string]int64, error)
	HashString(context.Context, string, string) string
	MulStringVal(context.Context, []*models.Pair) (map[string]int, error)
}

// RemoteConnector .
type RemoteConnector interface {
	Connect(context.Context) (io.ReadWriteCloser, error)
}

// TCPConnector .
type TCPConnector struct {
	addr string
}

// NewTCPConnector .
func NewTCPConnector(addr string) *TCPConnector {
	return &TCPConnector{addr: addr}
}

// Connect .
func (c *TCPConnector) Connect(ctx context.Context) (io.ReadWriteCloser, error) {
	var dialer net.Dialer

	conn, err := dialer.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return nil, fmt.Errorf("cant connect to remote server %w", err)
	}

	return conn, err
}

// TService .
type TService struct {
	DB        database.DB
	Connector RemoteConnector
	// remote    string
}

// NewTService .
func NewTService(db database.DB, connector RemoteConnector) *TService {
	return &TService{
		DB: db,
		// remote:    remoteServAddr,
		Connector: connector,
	}
}

// IncrementBy .
func (s *TService) IncrementBy(ctx context.Context, key string, val int64) (map[string]int64, error) {
	m := make(map[string]int64)

	i, err := s.DB.IncrementBy(ctx, key, val)
	if err != nil {
		return nil, err
	}

	m[key] = i
	return m, nil
}

// HashString .
func (s *TService) HashString(ctx context.Context, str, key string) string {

	fmt.Printf("str & key are %v, %v\n", str, key)

	h := hmac.New(sha512.New, []byte(key))

	// Never returns an error as doc says
	h.Write([]byte(str))

	bs := h.Sum(nil)
	return hex.EncodeToString(bs)

}

// MulStringVal .
func (s *TService) MulStringVal(ctx context.Context, pairs []*models.Pair) (map[string]int, error) {

	keys := make([]string, len(pairs))
	for i, v := range pairs {
		keys[i] = v.Key
	}

	str := MarshalMsg(pairs)

	conn, err := s.Connector.Connect(ctx)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	if _, err = conn.Write([]byte(str)); err != nil {
		return nil, fmt.Errorf("cant write to conn %w", err)
	}

	bs, err := ioutil.ReadAll(conn)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("cant read from conn %w", err)
	}

	m, err := UnmarshalMsg(keys, string(bs))
	if err != nil {
		return nil, err
	}

	return m, nil
}

// MarshalMsg .
func MarshalMsg(pairs []*models.Pair) string {
	var builder strings.Builder

	for _, v := range pairs {

		builder.WriteString(v.A)
		builder.WriteString(digitsep)
		builder.WriteString(v.B)
		builder.WriteString(pairsep)
	}

	builder.WriteString(pairsep)
	builder.WriteString(string(eof))

	return builder.String()
}

// UnmarshalMsg .
func UnmarshalMsg(keys []string, str string) (map[string]int, error) {

	m := make(map[string]int, 0)

	str = strings.TrimSuffix(str, pairsep+pairsep+string(eof))

	strs := strings.Split(str, "\r\n")

	if len(strs) != len(keys) {
		return nil, errors.New("not correct format")
	}

	for i, s := range strs {

		val, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("not correct format %w", err)
		}

		m[keys[i]] = val

	}

	return m, nil
}
