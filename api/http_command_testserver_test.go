package goxHttpApi

import (
	"context"
	"fmt"
	"github.com/devlibx/gox-base"
	"github.com/devlibx/gox-base/serialization"
	"github.com/devlibx/gox-base/test"
	"github.com/devlibx/gox-http/command"
	"github.com/devlibx/gox-http/testhelper"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func Test_Get_Success(t *testing.T) {
	cf, _ := test.MockCf(t)

	// Setup sample response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := gox.StringObjectMap{"status": "ok"}
		_, _ = fmt.Fprintln(w, serialization.StringifySuppressError(data, "{}"))
	}))
	defer ts.Close()

	// Read config and put the port to call
	config := command.Config{}
	err := serialization.ReadYamlFromString(testhelper.TestConfigWithRealServer, &config)
	assert.NoError(t, err)
	config.Servers["testServer"].Port, err = strconv.Atoi(strings.ReplaceAll(ts.URL, "http://127.0.0.1:", ""))
	assert.NoError(t, err)

	// Setup goHttp context
	goxHttpCtx, err := NewGoxHttpContext(cf, &config)
	assert.NoError(t, err)

	// Test 1 - Call http to get data
	ctx, ctxC := context.WithTimeout(context.Background(), 2*time.Second)
	defer ctxC()

	request := command.NewGoxRequestBuilder("delay_timeout_10").
		WithContentTypeJson().
		WithPathParam("id", 1).
		WithResponseBuilder(command.NewJsonToObjectResponseBuilder(&gox.StringObjectMap{})).
		Build()
	response, err := goxHttpCtx.Execute(ctx, "delay_timeout_10", request)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response.AsStringObjectMapOrEmpty().StringOrEmpty("status"))
}



func Test_Get_Timeout(t *testing.T) {
	cf, _ := test.MockCf(t)

	// Setup sample response with delay of 50 ms to fail this call
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		// data := gox.StringObjectMap{"status": "ok"}
		// _, _ = fmt.Fprintln(w, serialization.StringifySuppressError(data, "{}"))
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	// Read config and put the port to call
	config := command.Config{}
	err := serialization.ReadYamlFromString(testhelper.TestConfigWithRealServer, &config)
	assert.NoError(t, err)
	config.Servers["testServer"].Port, err = strconv.Atoi(strings.ReplaceAll(ts.URL, "http://127.0.0.1:", ""))
	assert.NoError(t, err)

	// Setup goHttp context
	goxHttpCtx, err := NewGoxHttpContext(cf, &config)
	assert.NoError(t, err)

	// Test 1 - Call http to get data
	ctx, ctxC := context.WithTimeout(context.Background(), 2*time.Second)
	defer ctxC()

	request := command.NewGoxRequestBuilder("delay_timeout_10").
		WithContentTypeJson().
		WithPathParam("id", 1).
		WithResponseBuilder(command.NewJsonToObjectResponseBuilder(&gox.StringObjectMap{})).
		Build()
	response, err := goxHttpCtx.Execute(ctx, "delay_timeout_10", request)
	assert.Error(t, err)
	fmt.Println(err)
	fmt.Println(response)
	// assert.Equal(t, "ok", response.AsStringObjectMapOrEmpty().StringOrEmpty("status"))
}
