package httpCommand

import (
	"context"
	"github.com/devlibx/gox-base"
	"github.com/devlibx/gox-base/errors"
	"github.com/devlibx/gox-base/serialization"
	"github.com/devlibx/gox-http/command"
	"github.com/go-resty/resty/v2"
	_ "github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"net"
	"net/http"
	"strings"
	"time"
)

type httpCommand struct {
	gox.CrossFunction
	server *command.Server
	api    *command.Api
	logger *zap.Logger
	client *resty.Client
}

func (h *httpCommand) ExecuteAsync(ctx context.Context, request *command.GoxRequest) chan *command.GoxResponse {
	responseChannel := make(chan *command.GoxResponse)
	go func() {
		if result, err := h.Execute(ctx, request); err != nil {
			responseChannel <- &command.GoxResponse{Err: err}
		} else {
			responseChannel <- result
		}
	}()
	return responseChannel
}

func (h *httpCommand) Execute(ctx context.Context, request *command.GoxRequest) (*command.GoxResponse, error) {
	h.logger.Debug("got request to execute", zap.Stringer("request", request))

	var response *resty.Response

	// Build request with all parameters
	r, err := h.buildRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// Create the url to call
	finalUrlToRequest := h.api.GetPath(h.server)
	h.logger.Debug("url to use", zap.String("url", finalUrlToRequest))

	switch strings.ToUpper(h.api.Method) {
	case "GET":
		response, err = r.Get(finalUrlToRequest)
	case "POST":
		response, err = r.Post(finalUrlToRequest)
	case "PUT":
		response, err = r.Put(finalUrlToRequest)
	case "DELETE":
		response, err = r.Delete(finalUrlToRequest)
	}

	if err != nil {
		responseObject := h.handleError(err)
		return responseObject, responseObject.Err
	} else {
		responseObject := h.processResponse(request, response)
		return responseObject, responseObject.Err
	}
}

func (h *httpCommand) buildRequest(ctx context.Context, request *command.GoxRequest) (*resty.Request, error) {
	r := h.client.R()
	r.SetContext(ctx)

	// Set header
	if request.Header != nil {
		for name, headers := range request.Header {
			for _, value := range headers {
				r.SetHeader(name, value)
			}
		}
	}

	// Auto set application/json as default
	if _, ok := request.Header["Content-Type"]; !ok {
		if _, ok := request.Header["content-type"]; !ok {
			r.SetHeader("content-type", "application/json")
		}
	}

	// Set query param
	if request.QueryParam != nil {
		for name, values := range request.QueryParam {
			for _, value := range values {
				r.SetQueryParam(name, value)
			}
		}
	}

	// Set path param
	if request.PathParam != nil {
		for name, values := range request.PathParam {
			for _, value := range values {
				r.SetPathParam(name, value)
			}
		}
	}

	if b, ok := request.Body.([]byte); ok {
		r.SetBody(b)
	} else if request.BodyProvider != nil {
		if b, err := request.BodyProvider.Body(request.Body); err == nil {
			r.SetBody(b)
		} else {
			return nil, &command.GoxHttpError{
				Err:        err,
				StatusCode: http.StatusInternalServerError,
				Message:    "failed to read body using body provider",
				ErrorCode:  command.ErrorCodeFailedToBuildRequest,
			}
		}
	} else {
		if b, err := serialization.Stringify(request.Body); err == nil {
			r.SetBody(b)
		} else {
			return nil, &command.GoxHttpError{
				Err:        err,
				StatusCode: http.StatusInternalServerError,
				Message:    "failed to read body using Stringify",
				ErrorCode:  command.ErrorCodeFailedToBuildRequest,
			}
		}
	}

	return r, nil
}

func (h *httpCommand) processResponse(request *command.GoxRequest, response *resty.Response) *command.GoxResponse {
	var processedResponse interface{}
	var err error

	if response.IsError() {

		if h.api.IsHttpCodeAcceptable(response.StatusCode()) {
			if request.ResponseBuilder != nil && response.Body() != nil {
				processedResponse, err = request.ResponseBuilder.Response(response.Body())
				if err != nil {
					return &command.GoxResponse{
						Body:       response.Body(),
						StatusCode: response.StatusCode(),
						Err: &command.GoxHttpError{
							Err:        errors.Wrap(err, "failed to create response using response builder"),
							StatusCode: response.StatusCode(),
							Message:    "failed to create response using response builder",
							ErrorCode:  "failed_to_build_response_using_response_builder",
							Body:       response.Body(),
						},
					}
				}
			}
		} else {
			return &command.GoxResponse{
				Body:       response.Body(),
				StatusCode: response.StatusCode(),
				Err: &command.GoxHttpError{
					Err:        errors.Wrap(err, "got response with server with error"),
					StatusCode: response.StatusCode(),
					Message:    "got response from server with error",
					ErrorCode:  "server_response_with_error",
					Body:       response.Body(),
				},
			}
		}

		return &command.GoxResponse{
			StatusCode: response.StatusCode(),
			Body:       response.Body(),
			Response:   processedResponse,
		}

	} else {

		if request.ResponseBuilder != nil && response.Body() != nil {
			processedResponse, err = request.ResponseBuilder.Response(response.Body())
			if err != nil {
				return &command.GoxResponse{
					Body:       response.Body(),
					StatusCode: response.StatusCode(),
					Err: &command.GoxHttpError{
						Err:        errors.Wrap(err, "failed to create response using response builder"),
						StatusCode: response.StatusCode(),
						Message:    "failed to create response using response builder",
						ErrorCode:  "failed_to_build_response_using_response_builder",
						Body:       response.Body(),
					},
				}
			}
		}

		return &command.GoxResponse{
			StatusCode: response.StatusCode(),
			Body:       response.Body(),
			Response:   processedResponse,
		}
	}
}

func (h *httpCommand) handleError(err error) *command.GoxResponse {
	var responseObject *command.GoxResponse

	// Timeout errors are handled here
	switch e := err.(type) {
	case net.Error:
		if e.Timeout() {
			responseObject = &command.GoxResponse{
				StatusCode: http.StatusRequestTimeout,
				Err: &command.GoxHttpError{
					Err:        e,
					StatusCode: http.StatusRequestTimeout,
					Message:    "request timeout on client",
					ErrorCode:  "request_timeout_on_client",
				},
			}
		}
	}

	// Not a timeout error
	if responseObject == nil {
		responseObject = &command.GoxResponse{
			StatusCode: http.StatusBadRequest,
			Err: &command.GoxHttpError{
				Err:        err,
				StatusCode: http.StatusBadRequest,
				Message:    "request failed on client",
				ErrorCode:  "request_failed_on_client",
			},
		}
	}

	return responseObject
}

func NewHttpCommand(cf gox.CrossFunction, server *command.Server, api *command.Api) (command.Command, error) {
	c := &httpCommand{
		CrossFunction: cf,
		server:        server,
		api:           api,
		logger:        cf.Logger().Named("goxHttp").Named(api.Name),
		client:        resty.New(),
	}
	c.client.SetTimeout(time.Duration(api.Timeout) * time.Millisecond)
	return c, nil
}
