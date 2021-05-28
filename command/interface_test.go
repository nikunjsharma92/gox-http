package command

import (
	"github.com/devlibx/gox-base/serialization"
	"github.com/devlibx/gox-http/testhelper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseConfig(t *testing.T) {
	config := Config{}
	err := serialization.ReadYamlFromString(testhelper.TestConfig, &config)
	assert.NoError(t, err)
	assert.True(t, len(config.Servers) > 0)
	assert.True(t, len(config.Apis) > 0)
}
