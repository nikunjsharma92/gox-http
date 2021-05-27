package goxHttpApi

import (
	"github.com/devlibx/gox-base"
	"github.com/devlibx/gox-base/errors"
	"github.com/devlibx/gox-http/command"
)

var ErrCommandNotRegisteredForApi = errors.New("api not found")

// Interface to be used by external clients
type GoxHttpContext interface {
	Execute(api string, request *command.GoxRequest) chan command.GoxResponse
}

// Create a new http context to be used
func NewGoxHttpContext(cf gox.CrossFunction, config *command.Config) (GoxHttpContext, error) {
	c := &goxHttpContextImpl{
		CrossFunction: cf,
		logger:        cf.Logger().Named("gox-http"),
		config:        config,
		commands:      map[string]command.Command{},
	}

	if err := c.setup(); err != nil {
		return nil, err
	}

	return c, nil
}
