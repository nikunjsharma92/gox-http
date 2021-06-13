package httpCommand

import (
	"context"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/devlibx/gox-base"
	"github.com/devlibx/gox-base/errors"
	"github.com/devlibx/gox-http/command"
	"go.uber.org/zap"
	"net/http"
)

var HystrixConfigMap = gox.StringObjectMap{}

type httpHystrixCommand struct {
	gox.CrossFunction
	logger             *zap.Logger
	command            command.Command
	hystrixCommandName string
	api                *command.Api
}

type result struct {
	response *command.GoxResponse
	err      error
}

func (h *httpHystrixCommand) Execute(ctx context.Context, request *command.GoxRequest) (*command.GoxResponse, error) {
	r := &result{}
	if err := hystrix.Do(h.hystrixCommandName, func() error {
		r.response, r.err = h.command.Execute(ctx, request)
		return r.err
	}, nil); err != nil {
		return r.response, h.errorCreator(err)
	} else {
		return r.response, r.err
	}
}

func (h *httpHystrixCommand) ExecuteAsync(ctx context.Context, request *command.GoxRequest) chan *command.GoxResponse {
	return nil
}

func (h *httpHystrixCommand) errorCreator(err error) error {
	if e, ok := err.(*command.GoxHttpError); ok {
		return e
	}

	switch e := err.(type) {
	case hystrix.CircuitError:
		if e == hystrix.ErrCircuitOpen {
			return &command.GoxHttpError{
				Err:        e,
				StatusCode: http.StatusBadRequest,
				Message:    "hystrix circuit open",
				ErrorCode:  "hystrix_circuit_open",
				Body:       nil,
			}
		} else if e == hystrix.ErrMaxConcurrency {
			return &command.GoxHttpError{
				Err:        e,
				StatusCode: http.StatusBadRequest,
				Message:    "hystrix rejected",
				ErrorCode:  "hystrix_rejected",
				Body:       nil,
			}
		} else if e == hystrix.ErrTimeout {
			return &command.GoxHttpError{
				Err:        e,
				StatusCode: http.StatusBadRequest,
				Message:    "hystrix timeout",
				ErrorCode:  "hystrix_timeout",
				Body:       nil,
			}
		} else {
			return &command.GoxHttpError{
				Err:        e,
				StatusCode: http.StatusBadRequest,
				Message:    "hystrix unknown ",
				ErrorCode:  "hystrix_unknown_error",
				Body:       nil,
			}
		}
	}

	return &command.GoxHttpError{
		Err:        err,
		StatusCode: http.StatusBadRequest,
		Message:    "unknown error",
		ErrorCode:  "unknown_error",
		Body:       nil,
	}
}

func NewHttpHystrixCommand(cf gox.CrossFunction, server *command.Server, api *command.Api) (command.Command, error) {

	hc, err := NewHttpCommand(cf, server, api)
	if err != nil {
		return nil, errors.Wrap(err, "failed to crate http command for %s", api.Name)
	}

	// name to register hystrix
	commandName := api.Name

	c := &httpHystrixCommand{
		CrossFunction:      cf,
		logger:             cf.Logger().Named("goxHttp").Named(api.Name),
		command:            hc,
		hystrixCommandName: commandName,
		api:                api,
	}

	// Set timeout + 10% delta
	timeout := api.Timeout

	// Add extra time to handle retry counts
	if api.RetryCount > 0 {
		timeout = timeout + (timeout * api.RetryCount) + api.InitialRetryWaitTimeMs
	}

	if timeout/10 <= 0 {
		timeout += 2
	} else {
		timeout += timeout / 10
	}

	// Inject setting - mostly used in testing
	config := HystrixConfigMap.StringObjectMapOrEmpty(api.Name)
	if config.IntOrZero("timeout") > 0 {
		timeout = config.IntOrZero("timeout")
	}

	hystrix.ConfigureCommand(commandName, hystrix.CommandConfig{
		Timeout:               timeout,
		MaxConcurrentRequests: api.Concurrency,
		ErrorPercentThreshold: 25,
	})

	return c, nil
}
