package pipeline

import (
	"testing"
	// "golang.org/x/net/context"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

func TestGenerateContent(t *testing.T) {
	b := &Builder{}
	expected_data, err := ioutil.ReadFile(`builder_test/pipeline01.yaml`)
	expected := Resources{}
	err = yaml.Unmarshal([]byte(expected_data), &expected)
	assert.NoError(t, err)
	result := b.GenerateDeploymentResources("dummy-proj", "pipeline01")
	assert.Equal(t, &expected, result)
}
