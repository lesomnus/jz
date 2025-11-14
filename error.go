//go:build js

package jz

import "syscall/js"

type RejectedError struct {
	Value js.Value
}

func (e RejectedError) Error() string {
	return Stringify(e.Value)
}
