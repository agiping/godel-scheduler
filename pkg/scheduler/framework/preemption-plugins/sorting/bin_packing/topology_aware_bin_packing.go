/*
Copyright 2023 The Godel Scheduler Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package topology

import (
	"sort"

	"k8s.io/apimachinery/pkg/runtime"

	framework "github.com/kubewharf/godel-scheduler/pkg/framework/api"
	"github.com/kubewharf/godel-scheduler/pkg/scheduler/framework/handle"
)

const TopologyAwareBinPackingName = "TopologyAwareBinPacking"

type TopologyAwareBinPacking struct{}

var _ framework.CandidatesSortingPlugin = &TopologyAwareBinPacking{}

func NewTopologyAwareBinPacking(_ runtime.Object, _ handle.PodFrameworkHandle) (framework.Plugin, error) {
	return &TopologyAwareBinPacking{}, nil
}

func (tabp *TopologyAwareBinPacking) Name() string {
	return TopologyAwareBinPackingName
}

func (tabp *TopologyAwareBinPacking) Compare(c1, c2 *framework.Candidate) int {
	score1 := calculateTopologyScore(c1.Node)
	score2 := calculateTopologyScore(c2.Node)

	if score1 > score2 {
		return 1
	} else if score1 < score2 {
		return -1
	}
	return 0
}

func calculateTopologyScore(node *framework.NodeInfo) int {
	topology := node.Topology()
	if topology == nil {
		return 0
	}

	score := 0
	numaNodes := topology.NumaNodes
	sort.Slice(numaNodes, func(i, j int) bool {
		return numaNodes[i].ID < numaNodes[j].ID
	})

	for i, numa := range numaNodes {
		cpuScore := calculateResourceScore(numa.CPUIds, node.AllocatableCPU())
		gpuScore := calculateResourceScore(numa.GPUIds, node.AllocatableGPU())

		// give lower NUMA ID a higher weight
		weight := len(numaNodes) - i
		score += (cpuScore + gpuScore) * weight
	}

	return score
}

func calculateResourceScore(allocatedIDs []int, allocatable int) int {
	if allocatable == 0 {
		return 0
	}
	return len(allocatedIDs) * 100 / allocatable
}
