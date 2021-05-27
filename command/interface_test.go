package command

import (
	"github.com/devlibx/gox-base/serialization"
	"github.com/devlibx/gox-http/testData"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseConfig(t *testing.T) {
	config := Config{}
	err := serialization.ReadYamlFromString(testData.TestConfig, &config)
	assert.NoError(t, err)
	assert.True(t, len(config.Servers) > 0)
	assert.True(t, len(config.Apis) > 0)
}
