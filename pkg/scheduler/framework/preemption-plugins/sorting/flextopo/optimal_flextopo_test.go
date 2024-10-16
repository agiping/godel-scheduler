package flextopo

import (
	"testing"

	"github.com/stretchr/testify/assert"

	framework "github.com/kubewharf/godel-scheduler/pkg/framework/api"
)

func TestNewOptimalFlextopo(t *testing.T) {
	plugin, err := NewOptimalFlextopo(nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, plugin)
	assert.IsType(t, &OptimalFlextopo{}, plugin)
}

func TestOptimalFlextopoName(t *testing.T) {
	plugin := &OptimalFlextopo{}
	assert.Equal(t, OptimalFlextopoName, plugin.Name())
}

func TestOptimalFlextopoCompare(t *testing.T) {
	plugin := &OptimalFlextopo{}
	c1 := &framework.Candidate{}
	c2 := &framework.Candidate{}

	result := plugin.Compare(c1, c2)
	assert.Equal(t, 0, result, "expected comparison result to be 0")
}

func TestOptimalFlextopoImplementsCandidatesSortingPlugin(t *testing.T) {
	var _ framework.CandidatesSortingPlugin = &OptimalFlextopo{}
}
