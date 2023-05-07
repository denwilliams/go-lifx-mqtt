package mqtt

import (
	"fmt"
)

type Command struct {
	Brightness  *uint16 `json:"brightness"`
	Color       *string `json:"color"`
	Temperature *uint16 `json:"temp"`
	Duration    uint32  `json:"duration"`
}

func safeUint16(s *uint16) string {
	if s == nil {
		return "(nil)"
	}
	return fmt.Sprintf("%d", *s)
}

func safeString(s *string) string {
	if s == nil {
		return "(nil)"
	}
	return (*s)
}

func (c *Command) String() string {
	return fmt.Sprintf("brightness=%s color=%s temperature=%s duration=%d", safeUint16(c.Brightness), safeString(c.Color), safeUint16(c.Temperature), c.Duration)
}

type CommandHandler interface {
	HandleCommand(id string, command *Command) error
}
