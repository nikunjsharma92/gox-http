package httpCommand

import (
	"fmt"
	"github.com/devlibx/gox-base"
	"github.com/devlibx/gox-base/test"
	"github.com/devlibx/gox-http/command"
	"github.com/devlibx/gox-http/testData"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHttpCommand(t *testing.T) {
	cf, _ := test.MockCf(t)

	config := command.Config{}
	err := testData.GetTestConfig(&config)
	assert.NoError(t, err)

	server, err := config.FindServerByName("jsonplaceholder")
	assert.NoError(t, err)

	api, err := config.FindApiByName("getPosts")
	assert.NoError(t, err)

	httpCmd, err := NewHttpCommand(cf, server, api)
	assert.NoError(t, err)

	result := <-httpCmd.Execute(&command.GoxRequest{
		PathParam: map[string][]string{"id": {"1"}},
		ResponseBuilder: command.NewFunctionBasedResponseBuilder(func(data []byte) (interface{}, error) {
			return gox.StringObjectMapFromString(string(data))
		}),
	})
	assert.NoError(t, result.Err)
	fmt.Println(result.Response)
}
