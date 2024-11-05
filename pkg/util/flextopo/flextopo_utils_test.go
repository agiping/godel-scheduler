package flextopo

import (
	"testing"

	flextopov1alpha1 "github.com/agiping/flextopo-api/pkg/apis/flextopo/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestGetPodNUMAAndSockets1(t *testing.T) {
	// Create a FlexTopo example
	flexTopo := &flextopov1alpha1.FlexTopo{
		Spec: flextopov1alpha1.FlexTopoSpec{
			Nodes: []flextopov1alpha1.FlexTopoNode{
				{ID: "core1", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod1"}},
				{ID: "numa1", Type: "NUMANode"},
				{ID: "socket1", Type: "Socket"},
			},
			Edges: []flextopov1alpha1.FlexTopoEdge{
				{Source: "numa1", Target: "core1", Type: "contains"},
				{Source: "socket1", Target: "numa1", Type: "contains"},
			},
		},
	}

	// test cases
	tests := []struct {
		podNames []string
		expected PodTopologyInfo
	}{
		{
			podNames: []string{"pod1"},
			expected: PodTopologyInfo{
				"pod1": {
					"numas":   []string{"numa1"},
					"sockets": []string{"socket1"},
				},
			},
		},
	}

	for _, tt := range tests {
		result := GetPodNUMAAndSockets(flexTopo, tt.podNames)
		if len(result) != len(tt.expected) {
			t.Errorf("expected %v, got %v", tt.expected, result)
		}
		for podName, topology := range tt.expected {
			if len(result[podName]["numas"]) != len(topology["numas"]) || len(result[podName]["sockets"]) != len(topology["sockets"]) {
				t.Errorf("expected %v, got %v", topology, result[podName])
			}
		}
	}
}

func TestGetPodNUMAAndSockets2(t *testing.T) {
	flexTopo2 := &flextopov1alpha1.FlexTopo{
		Spec: flextopov1alpha1.FlexTopoSpec{
			Nodes: []flextopov1alpha1.FlexTopoNode{
				{ID: "core1", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod3"}},
				{ID: "core2", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod3"}},
				{ID: "core3", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod3"}},
				{ID: "core4", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod3"}},
				{ID: "core5", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod4"}},
				{ID: "core6", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod4"}},
				{ID: "core7", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod4"}},
				{ID: "core8", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod4"}},
				{ID: "numa0", Type: "NUMANode"},
				{ID: "numa1", Type: "NUMANode"},
				{ID: "socket0", Type: "Socket"},
				{ID: "socket1", Type: "Socket"},
			},
			Edges: []flextopov1alpha1.FlexTopoEdge{
				{Source: "numa0", Target: "core1", Type: "contains"},
				{Source: "numa0", Target: "core2", Type: "contains"},
				{Source: "numa0", Target: "core3", Type: "contains"},
				{Source: "numa0", Target: "core4", Type: "contains"},
				{Source: "numa1", Target: "core5", Type: "contains"},
				{Source: "numa1", Target: "core6", Type: "contains"},
				{Source: "numa1", Target: "core7", Type: "contains"},
				{Source: "numa1", Target: "core8", Type: "contains"},
				{Source: "socket0", Target: "numa0", Type: "contains"},
				{Source: "socket0", Target: "numa1", Type: "contains"},
			},
		},
	}

	tests := []struct {
		podNames []string
		expected PodTopologyInfo
	}{
		{
			podNames: []string{"pod3", "pod4"},
			expected: PodTopologyInfo{
				"pod3": {
					"numas":   []string{"numa0"},
					"sockets": []string{"socket0"},
				},
				"pod4": {
					"numas":   []string{"numa1"},
					"sockets": []string{"socket0"},
				},
			},
		},
	}

	for _, tt := range tests {
		result := GetPodNUMAAndSockets(flexTopo2, tt.podNames)
		if len(result) != len(tt.expected) {
			t.Errorf("expected %v, got %v", tt.expected, result)
		}
		for podName, topology := range tt.expected {
			if len(result[podName]["numas"]) != len(topology["numas"]) || len(result[podName]["sockets"]) != len(topology["sockets"]) {
				t.Errorf("expected %v, got %v", topology, result[podName])
			}
		}
		assert.Equal(t, tt.expected, result)
	}
}
