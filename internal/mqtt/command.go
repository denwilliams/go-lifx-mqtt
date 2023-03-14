package mqtt

import (
	"fmt"
)

type Command struct {
	Brightness  uint16 `json:"brightness"`
	Color       string `json:"color"`
	Temperature uint16 `json:"temp"`
	Duration    uint32 `json:"duration"`
}

func (c *Command) String() string {
	return fmt.Sprintf("brightness:%d color:%s temperature:%d", c.Brightness, c.Color, c.Temperature)
}

type CommandHandler interface {
	HandleCommand(id string, command *Command) error
}
