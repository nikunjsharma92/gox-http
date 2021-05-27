package goxHttpApi

import (
	"github.com/devlibx/gox-base"
	"github.com/devlibx/gox-base/errors"
	"github.com/devlibx/gox-http/command"
	httpCommand "github.com/devlibx/gox-http/command/http"
	"go.uber.org/zap"
)

// Implementation of http context
type goxHttpContextImpl struct {
	gox.CrossFunction
	logger   *zap.Logger
	config   *command.Config
	commands map[string]command.Command
}

// Internal setup method
func (g *goxHttpContextImpl) setup() error {
	g.config.SetupDefaults()

	for apiName, api := range g.config.Apis {

		// Find the server used in this API
		server, err := g.config.FindServerByName(api.Server)
		if err != nil {
			return errors.Wrap(err, "failed to create http command (server not found): api=%s", apiName)
		}

		// Create http command for this API
		cmd, err := httpCommand.NewHttpCommand(g.CrossFunction, server, api)
		if err != nil {
			return errors.Wrap(err, "failed to create http command: api=%s", apiName)
		}

		// Store this http command to use
		g.commands[apiName] = cmd

	}
	return nil
}

// Execute a request
func (g *goxHttpContextImpl) Execute(api string, request *command.GoxRequest) chan command.GoxResponse {

	if cmd, ok := g.commands[api]; !ok {
		// Api command not found - return a error channel
		resultChannel := make(chan command.GoxResponse, 1)
		resultChannel <- command.GoxResponse{
			Err: errors.Wrap(ErrCommandNotRegisteredForApi, "command to execute not found: name=%s", api),
		}
		close(resultChannel)
		return resultChannel
	} else {
		_ = cmd

		resultChannel := make(chan command.GoxResponse, 2)
		close(resultChannel)
		return resultChannel
	}
}
