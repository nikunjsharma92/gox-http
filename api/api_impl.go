package goxHttpApi

import (
	"context"
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

func (g *goxHttpContextImpl) Execute(ctx context.Context, api string, request *command.GoxRequest) (*command.GoxResponse, error) {
	if cmd, ok := g.commands[api]; !ok {
		return nil, errors.Wrap(ErrCommandNotRegisteredForApi, "command to execute not found: name=%s", api)
	} else {
		return cmd.Execute(ctx, request)
	}
}

func (g *goxHttpContextImpl) ExecuteAsync(ctx context.Context, api string, request *command.GoxRequest) chan *command.GoxResponse {
	panic("implement me")
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
		var cmd command.Command
		if api.DisableHystrix {
			cmd, err = httpCommand.NewHttpCommand(g.CrossFunction, server, api)
		} else {
			cmd, err = httpCommand.NewHttpHystrixCommand(g.CrossFunction, server, api)
		}
		if err != nil {
			return errors.Wrap(err, "failed to create http command: api=%s", apiName)
		}

		// Store this http command to use
		g.commands[apiName] = cmd

	}
	return nil
}
