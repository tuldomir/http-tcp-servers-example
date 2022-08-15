package models

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// IncrMsgIn .
type IncrMsgIn struct {
	Key string `json:"key"`
	Val int64  `json:"val"`
}

// Validate .
func (msg IncrMsgIn) Validate() error {
	return validation.ValidateStruct(&msg,
		validation.Field(&msg.Key, validation.Required, validation.Length(1, 20)),
		validation.Field(&msg.Val, validation.Required),
	)
}

// IncrMsgOut .
type IncrMsgOut struct {
	Res int64 `json:"res"`
}

// HashMsgIn .
type HashMsgIn struct {
	S   string `json:"s,omitempty"`
	Key string `json:"key,omitempty"`
}

// Validate .
func (msg HashMsgIn) Validate() error {
	return validation.ValidateStruct(&msg,
		validation.Field(&msg.S, validation.Required, validation.Length(1, 20)),
		validation.Field(&msg.Key, validation.Required, validation.Length(1, 20)),
	)
}

// Pair .
type Pair struct {
	A   string `josn:"a"`
	B   string `json:"b"`
	Key string `josn:"key"`
}

// Validate .
func (p Pair) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.A, validation.Required),
		validation.Field(&p.B, validation.Required),
		validation.Field(&p.Key, validation.Required, validation.Length(1, 20)),
	)
}
