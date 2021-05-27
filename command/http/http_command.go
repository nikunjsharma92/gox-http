package httpCommand

import (
	"context"
	"github.com/devlibx/gox-base"
	"github.com/devlibx/gox-base/errors"
	"github.com/devlibx/gox-http/command"
	"github.com/go-resty/resty/v2"
	_ "github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"net/http"
	"strings"
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
		responseObject := h.processError(request, response, err)
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
			return nil, errors.Wrap(err, "failed to set body")
		}
	}

	return r, nil
}

func (h *httpCommand) processError(request *command.GoxRequest, response *resty.Response, errorFromCall error) *command.GoxResponse {
	if errorFromCall != nil && response != nil {
		if h.api.IsHttpCodeAcceptable(response.StatusCode()) {
			return h.processResponse(request, response)
		} else {
			return &command.GoxResponse{
				Body:       response.Body(),
				StatusCode: response.StatusCode(),
				Err:        errors.Wrap(errorFromCall, "got error from server with response"),
			}
		}
	} else if errorFromCall != nil {
		if h.api.IsHttpCodeAcceptable(http.StatusInternalServerError) {
			return h.processResponse(request, response)
		} else {
			return &command.GoxResponse{
				StatusCode: http.StatusInternalServerError,
				Err:        errors.Wrap(errorFromCall, "got error from server without response"),
			}
		}
	}
	return nil
}

func (h *httpCommand) processResponse(request *command.GoxRequest, response *resty.Response) *command.GoxResponse {

	var processedResponse interface{}
	var errorInBuildingResponse error
	if request.ResponseBuilder != nil && response.Body() != nil {
		processedResponse, errorInBuildingResponse = request.ResponseBuilder.Response(response.Body())
	}

	if errorInBuildingResponse != nil {
		return &command.GoxResponse{
			StatusCode: response.StatusCode(),
			Body:       response.Body(),
			Err:        errors.Wrap(errorInBuildingResponse, "got error from server without response"),
		}
	} else {
		return &command.GoxResponse{
			StatusCode: response.StatusCode(),
			Body:       response.Body(),
			Response:   processedResponse,
			Err:        errors.Wrap(errorInBuildingResponse, "got error from server without response"),
		}
	}
}

func NewHttpCommand(cf gox.CrossFunction, server *command.Server, api *command.Api) (command.Command, error) {
	c := &httpCommand{
		CrossFunction: cf,
		server:        server,
		api:           api,
		logger:        cf.Logger().Named("goxHttp").Named(api.Name),
		client:        resty.New(),
	}
	return c, nil
}
