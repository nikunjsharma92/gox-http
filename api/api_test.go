package goxHttpApi

import (
	"errors"
	"github.com/devlibx/gox-base/test"
	"github.com/devlibx/gox-http/command"
	"github.com/devlibx/gox-http/testData"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGoxHttpContext_WithNonExistingApiName(t *testing.T) {
	cf, _ := test.MockCf(t)

	config := command.Config{}
	err := testData.GetTestConfig(&config)
	assert.NoError(t, err)

	goxHttpCtx, err := NewGoxHttpContext(cf, &config)
	assert.NoError(t, err)

	result := <-goxHttpCtx.Execute("badName", nil)
	assert.Error(t, result.Err)
	assert.True(t, errors.Is(result.Err, ErrCommandNotRegisteredForApi))
}
