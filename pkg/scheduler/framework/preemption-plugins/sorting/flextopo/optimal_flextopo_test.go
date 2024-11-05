package flextopo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flextopov1alpha1 "github.com/agiping/flextopo-api/pkg/apis/flextopo/v1alpha1"
	framework "github.com/kubewharf/godel-scheduler/pkg/framework/api"
	ftopoutil "github.com/kubewharf/godel-scheduler/pkg/util/flextopo"
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
	c1 := &framework.Candidate{
		FlexTopo: &flextopov1alpha1.FlexTopo{},
		Name:     "node1",
		Victims: &framework.Victims{
			Pods: []*v1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod2"}},
			},
		},
	}
	c2 := &framework.Candidate{
		FlexTopo: &flextopov1alpha1.FlexTopo{},
		Name:     "node2",
		Victims: &framework.Victims{
			Pods: []*v1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod3"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod4"}},
			},
		},
	}

	result := plugin.Compare(c1, c2)
	assert.Equal(t, 0, result, "expected comparison result to be 0")
}

func TestOptimalFlextopoImplementsCandidatesSortingPlugin(t *testing.T) {
	var _ framework.CandidatesSortingPlugin = &OptimalFlextopo{}
}

func TestGetAlignmentScore(t *testing.T) {
	// Case 1: the worest case, victims in different numa nodes and crossing sockets
	flexTopo1 := &flextopov1alpha1.FlexTopo{
		Spec: flextopov1alpha1.FlexTopoSpec{
			Nodes: []flextopov1alpha1.FlexTopoNode{
				{ID: "core1", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod1"}},
				{ID: "core2", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod1"}},
				{ID: "core3", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod1"}},
				{ID: "core4", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod1"}},
				{ID: "core5", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod2"}},
				{ID: "core6", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod2"}},
				{ID: "core7", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod2"}},
				{ID: "core8", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod2"}},
				{ID: "numa1", Type: "NUMANode"},
				{ID: "numa2", Type: "NUMANode"},
				{ID: "socket0", Type: "Socket"},
				{ID: "socket1", Type: "Socket"},
			},
			Edges: []flextopov1alpha1.FlexTopoEdge{
				{Source: "numa1", Target: "core1", Type: "contains"},
				{Source: "numa1", Target: "core2", Type: "contains"},
				{Source: "numa1", Target: "core3", Type: "contains"},
				{Source: "numa1", Target: "core4", Type: "contains"},
				{Source: "socket0", Target: "numa1", Type: "contains"},
				{Source: "numa2", Target: "core5", Type: "contains"},
				{Source: "numa2", Target: "core6", Type: "contains"},
				{Source: "numa2", Target: "core7", Type: "contains"},
				{Source: "numa2", Target: "core8", Type: "contains"},
				{Source: "socket1", Target: "numa2", Type: "contains"},
			},
		},
	}

	expected := ftopoutil.PodTopologyInfo{
		"pod1": {
			"numas":   []string{"numa1"},
			"sockets": []string{"socket0"},
		},
		"pod2": {
			"numas":   []string{"numa2"},
			"sockets": []string{"socket1"},
		},
	}
	result := ftopoutil.GetPodNUMAAndSockets(flexTopo1, []string{"pod1", "pod2"})
	assert.Equal(t, expected, result)

	candidate := &framework.Candidate{
		FlexTopo: flexTopo1,
		Name:     "node1",
		Victims: &framework.Victims{
			Pods: []*v1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod2"}},
			},
		},
	}
	expectedScore := 7
	score := getAlignmentScore(candidate)
	assert.Equal(t, expectedScore, score)

	// Case 2: acceptable case, victims in different numa nodes but in the same socket
	flexTopo2 := &flextopov1alpha1.FlexTopo{
		Spec: flextopov1alpha1.FlexTopoSpec{
			Nodes: []flextopov1alpha1.FlexTopoNode{
				{ID: "core1", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod1"}},
				{ID: "core2", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod1"}},
				{ID: "core3", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod1"}},
				{ID: "core4", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod1"}},
				{ID: "core5", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod2"}},
				{ID: "core6", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod2"}},
				{ID: "core7", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod2"}},
				{ID: "core8", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod2"}},
				{ID: "numa1", Type: "NUMANode"},
				{ID: "numa2", Type: "NUMANode"},
				{ID: "socket0", Type: "Socket"},
			},
			Edges: []flextopov1alpha1.FlexTopoEdge{
				{Source: "numa1", Target: "core1", Type: "contains"},
				{Source: "numa1", Target: "core2", Type: "contains"},
				{Source: "numa1", Target: "core3", Type: "contains"},
				{Source: "numa1", Target: "core4", Type: "contains"},
				{Source: "numa2", Target: "core5", Type: "contains"},
				{Source: "numa2", Target: "core6", Type: "contains"},
				{Source: "numa2", Target: "core7", Type: "contains"},
				{Source: "numa2", Target: "core8", Type: "contains"},
				{Source: "socket0", Target: "numa1", Type: "contains"},
				{Source: "socket0", Target: "numa2", Type: "contains"},
			},
		},
	}
	expected = ftopoutil.PodTopologyInfo{
		"pod1": {
			"numas":   []string{"numa1"},
			"sockets": []string{"socket0"},
		},
		"pod2": {
			"numas":   []string{"numa2"},
			"sockets": []string{"socket0"},
		},
	}
	result = ftopoutil.GetPodNUMAAndSockets(flexTopo2, []string{"pod1", "pod2"})
	assert.Equal(t, expected, result)

	candidate = &framework.Candidate{
		FlexTopo: flexTopo2,
		Name:     "node2",
		Victims: &framework.Victims{
			Pods: []*v1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod2"}},
			},
		},
	}
	expectedScore = 12
	score = getAlignmentScore(candidate)
	assert.Equal(t, expectedScore, score)

	// Case 3: the best case, victims in the same numa node
	flexTopo3 := &flextopov1alpha1.FlexTopo{
		Spec: flextopov1alpha1.FlexTopoSpec{
			Nodes: []flextopov1alpha1.FlexTopoNode{
				{ID: "core1", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod1"}},
				{ID: "core2", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod1"}},
				{ID: "core3", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod1"}},
				{ID: "core4", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod1"}},
				{ID: "core5", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod2"}},
				{ID: "core6", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod2"}},
				{ID: "core7", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod2"}},
				{ID: "core8", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod2"}},
				{ID: "numa1", Type: "NUMANode"},
				{ID: "socket0", Type: "Socket"},
			},
			Edges: []flextopov1alpha1.FlexTopoEdge{
				{Source: "numa1", Target: "core1", Type: "contains"},
				{Source: "numa1", Target: "core2", Type: "contains"},
				{Source: "numa1", Target: "core3", Type: "contains"},
				{Source: "numa1", Target: "core4", Type: "contains"},
				{Source: "numa1", Target: "core5", Type: "contains"},
				{Source: "numa1", Target: "core6", Type: "contains"},
				{Source: "numa1", Target: "core7", Type: "contains"},
				{Source: "numa1", Target: "core8", Type: "contains"},
				{Source: "socket0", Target: "numa1", Type: "contains"},
			},
		},
	}
	expected = ftopoutil.PodTopologyInfo{
		"pod1": {
			"numas":   []string{"numa1"},
			"sockets": []string{"socket0"},
		},
		"pod2": {
			"numas":   []string{"numa1"},
			"sockets": []string{"socket0"},
		},
	}
	result = ftopoutil.GetPodNUMAAndSockets(flexTopo3, []string{"pod1", "pod2"})
	assert.Equal(t, expected, result)

	candidate = &framework.Candidate{
		FlexTopo: flexTopo3,
		Name:     "node3",
		Victims: &framework.Victims{
			Pods: []*v1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod2"}},
			},
		},
	}
	expectedScore = 15
	score = getAlignmentScore(candidate)
	assert.Equal(t, expectedScore, score)
}

func TestCompare(t *testing.T) {
	plugin := &OptimalFlextopo{}

	flexTopo1 := &flextopov1alpha1.FlexTopo{
		Spec: flextopov1alpha1.FlexTopoSpec{
			Nodes: []flextopov1alpha1.FlexTopoNode{
				{ID: "core1", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod1"}},
				{ID: "core2", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod1"}},
				{ID: "core3", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod1"}},
				{ID: "core4", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod1"}},
				{ID: "core35", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod2"}},
				{ID: "core36", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod2"}},
				{ID: "core37", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod2"}},
				{ID: "core38", Type: "CPUCore", Attributes: map[string]string{"usedBy": "pod2"}},
				{ID: "numa0", Type: "NUMANode"},
				{ID: "numa5", Type: "NUMANode"},
				{ID: "socket0", Type: "Socket"},
				{ID: "socket1", Type: "Socket"},
			},
			Edges: []flextopov1alpha1.FlexTopoEdge{
				{Source: "numa0", Target: "core1", Type: "contains"},
				{Source: "numa0", Target: "core2", Type: "contains"},
				{Source: "numa0", Target: "core3", Type: "contains"},
				{Source: "numa0", Target: "core4", Type: "contains"},
				{Source: "socket0", Target: "numa0", Type: "contains"},
				{Source: "numa5", Target: "core35", Type: "contains"},
				{Source: "numa5", Target: "core36", Type: "contains"},
				{Source: "numa5", Target: "core37", Type: "contains"},
				{Source: "numa5", Target: "core38", Type: "contains"},
				{Source: "socket1", Target: "numa5", Type: "contains"},
			},
		},
	}

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

	expectedTopology1 := ftopoutil.PodTopologyInfo{
		"pod1": {
			"numas":   []string{"numa0"},
			"sockets": []string{"socket0"},
		},
		"pod2": {
			"numas":   []string{"numa5"},
			"sockets": []string{"socket1"},
		},
	}
	result1 := ftopoutil.GetPodNUMAAndSockets(flexTopo1, []string{"pod1", "pod2"})
	assert.Equal(t, expectedTopology1, result1)

	expectedTopology2 := ftopoutil.PodTopologyInfo{
		"pod3": {
			"numas":   []string{"numa0"},
			"sockets": []string{"socket0"},
		},
		"pod4": {
			"numas":   []string{"numa1"},
			"sockets": []string{"socket0"},
		},
	}
	result2 := ftopoutil.GetPodNUMAAndSockets(flexTopo2, []string{"pod3", "pod4"})
	assert.Equal(t, expectedTopology2, result2)

	c1 := &framework.Candidate{
		FlexTopo: flexTopo1,
		Victims: &framework.Victims{
			Pods: []*v1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod2"}},
			},
		},
	}
	c2 := &framework.Candidate{
		FlexTopo: flexTopo2,
		Victims: &framework.Victims{
			Pods: []*v1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod4"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod3"}},
			},
		},
	}
	expectedScore1 := 7
	expectedScore2 := 12
	assert.Equal(t, expectedScore1, getAlignmentScore(c1))
	assert.Equal(t, expectedScore2, getAlignmentScore(c2))

	result := plugin.Compare(c1, c2)
	assert.Equal(t, 1, result)
}
