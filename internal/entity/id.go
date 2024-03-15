package entity

import (
	"fmt"
	"strings"
)

type ID string

func (i ID) Type() string {
	return strings.Split(string(i), ":")[0]
}

func (i ID) ID() string {
	return strings.Split(string(i), ":")[1]
}
func (i ID) String() string {
	return string(i)
}

func NewID[T any](entityType string, id T) ID {
	return ID(fmt.Sprintf("%s:%v", entityType, id))
}
