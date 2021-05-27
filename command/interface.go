package command

import (
	"context"
	"fmt"
	"net/http"
)

// List of all servers
type Servers map[string]*Server

// Defines a single server
type Server struct {
	Name                     string
	Host                     string `yaml:"host"`
	Port                     int    `yaml:"port"`
	Https                    bool   `yaml:"https"`
	ConnectTimeout           int    `yaml:"connect_timeout"`
	ConnectionRequestTimeout int    `yaml:"connection_request_timeout"`
}

// List of all APIs
type Apis map[string]*Api

// A single API
type Api struct {
	Name            string
	Method          string `yaml:"method"`
	Path            string `yaml:"path"`
	Server          string `yaml:"server"`
	Timeout         int    `yaml:"timeout"`
	Concurrency     int    `yaml:"concurrency"`
	QueueSize       int    `yaml:"queue_size"`
	Async           bool   `yaml:"async"`
	AcceptableCodes string `yaml:"acceptable_codes"`
	acceptableCodes []int
}

type Config struct {
	Servers Servers `yaml:"servers"`
	Apis    Apis    `yaml:"apis"`
}

// ------------------------------------------------------ Request/Response ---------------------------------------------

type MultivaluedMap map[string][]string

type BodyProvider interface {
	Body(object interface{}) ([]byte, error)
}

type ResponseBuilder interface {
	Response(data []byte) (interface{}, error)
}

type GoxRequest struct {
	Header          http.Header
	PathParam       MultivaluedMap
	QueryParam      MultivaluedMap
	Body            interface{}
	BodyProvider    BodyProvider
	ResponseBuilder ResponseBuilder
}

type GoxResponse struct {
	Body       []byte
	Response   interface{}
	StatusCode int
	Err        error
}

type Command interface {
	Execute(ctx context.Context, request *GoxRequest) (*GoxResponse, error)
	ExecuteAsync(ctx context.Context, request *GoxRequest) chan *GoxResponse
}

func (req *GoxRequest) String() string {
	return fmt.Sprintf("")
}
