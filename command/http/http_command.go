package httpCommand

import (
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

func (h *httpCommand) Execute(request *command.GoxRequest) chan command.GoxResponse {
	h.logger.Debug("got request to execute", zap.Stringer("request", request))
	responseChannel := make(chan command.GoxResponse, 1)

	var err error
	var response *resty.Response

	// Build request with all parameters
	r, err := h.buildRequest(request)

	// We got error stop here
	if err != nil {
		responseChannel <- command.GoxResponse{Err: err}
		close(responseChannel)
		return responseChannel
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

	if err != nil && response != nil {
		responseChannel <- command.GoxResponse{
			Body:       response.Body(),
			StatusCode: response.StatusCode(),
			Err:        errors.Wrap(err, "got error from server with response"),
		}
	} else if err != nil {
		responseChannel <- command.GoxResponse{
			StatusCode: http.StatusInternalServerError,
			Err:        errors.Wrap(err, "got error from server without response"),
		}
	} else {

		var processedResponse interface{}
		var errorInBuildingResponse error
		if request.ResponseBuilder != nil && response.Body() != nil {
			processedResponse, errorInBuildingResponse = request.ResponseBuilder.Response(response.Body())
		}

		if errorInBuildingResponse != nil {
			responseChannel <- command.GoxResponse{
				StatusCode: response.StatusCode(),
				Body:       response.Body(),
				Err:        errors.Wrap(errorInBuildingResponse, "got error from server without response"),
			}
		} else {
			responseChannel <- command.GoxResponse{
				StatusCode: response.StatusCode(),
				Body:       response.Body(),
				Response:   processedResponse,
				Err:        errors.Wrap(errorInBuildingResponse, "got error from server without response"),
			}
		}
	}

	close(responseChannel)
	return responseChannel
}

func (h *httpCommand) buildRequest(request *command.GoxRequest) (*resty.Request, error) {
	r := h.client.R()

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
