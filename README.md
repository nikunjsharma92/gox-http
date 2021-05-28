## Gox Http
Gox Http provides utility to call a http endpoint. It provides following:

1. Define all endpoint and api config in configuration file
2. Circuit breaker using Hytrix

#### How to use
Given below is a example on how to use this liberary
```go
package main

import (
	"context"
	"fmt"
	"github.com/devlibx/gox-base"
	"github.com/devlibx/gox-base/serialization"
	goxHttpApi "github.com/devlibx/gox-http/api"
	"github.com/devlibx/gox-http/command"
	"log"
)

// Here you can define your own configuration
// We have used "jsonplaceholder" as a test server. A api "getPosts" is defined which uses "server=jsonplaceholder"
var httpConfig = `
servers:
  jsonplaceholder:
    host: jsonplaceholder.typicode.com
    port: 443
    https: true
    connect_timeout: 1000
    connection_request_timeout: 1000
  testServer:
    host: localhost
    port: 9123

apis:
  getPosts:
    method: GET
    path: /posts/{id}
    server: jsonplaceholder
    timeout: 1000
    acceptable_codes: 200,201
  delay_timeout_10:
    path: /delay
    server: testServer
    timeout: 10
    concurrency: 3 
`

func main() {

	cf := gox.NewCrossFunction()

	// Read config and
	config := command.Config{}
	err := serialization.ReadYamlFromString(httpConfig, &config)
	if err != nil {
		log.Println("got error in reading config", err)
		return
	}

	// Setup goHttp context
	goxHttpCtx, err := goxHttpApi.NewGoxHttpContext(cf, &config)
	if err != nil {
		log.Println("got error in creating gox http context config", err)
		return
	}

	// Make a http call and get the result
	// 	ResponseBuilder - this is used to convert json response to your custom object
	//
	//  The following interface can be implemented to convert from bytes to the desired output.
	//  response.Response will hold the object which is returned from  ResponseBuilder
	//
	//	type ResponseBuilder interface {
	//		Response(data []byte) (interface{}, error)
	//	}
	request := command.NewGoxRequestBuilder("getPosts").
		WithContentTypeJson().
		WithPathParam("id", 1).
		WithResponseBuilder(command.NewJsonToObjectResponseBuilder(&gox.StringObjectMap{})).
		Build()
	response, err := goxHttpCtx.Execute(context.Background(), "getPosts", request)
	if err != nil {

		// Error details can be extracted from *command.GoxHttpError
		if goxError, ok := err.(*command.GoxHttpError); ok {
			if goxError.Is5xx() {
				fmt.Println("got 5xx error")
			} else if goxError.Is4xx() {
				fmt.Println("got 5xx error")
			} else if goxError.IsBadRequest() {
				fmt.Println("got bad request error")
			}
		} else {
			fmt.Println("got unknown error")
		}

	} else {
		fmt.Println(serialization.Stringify(response.Response))
		// {some json response ...}
	}
}
```