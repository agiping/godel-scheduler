/*
Copyright 2024 The FlexTopo Authors.

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

package flextopo

import (
	"math"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"

	framework "github.com/kubewharf/godel-scheduler/pkg/framework/api"
	"github.com/kubewharf/godel-scheduler/pkg/scheduler/framework/handle"
	ftopoutil "github.com/kubewharf/godel-scheduler/pkg/util/flextopo"
	podutil "github.com/kubewharf/godel-scheduler/pkg/util/pod"
)

const (
	OptimalFlextopoName string = "OptimalFlextopo"
	// this is the weight of the priority and topology alignment score, i.e., alpha in the formula:
	// score = alpha * 1/priority_score + (1 - alpha) * topology_score,
	// 0 <= alpha <= 1
	// Note that the value of priority_score and topology_score are both normalized
	// to [0, 1]. We only care about the relative order of the scores.
	PriorityTopologyAlignmentAlpha float64 = 0
	// topological alignment score level
	MaxTopologyScore       int = 15
	LossOfDiffNumasScore   int = 3 // if in different numa nodes, lose 3 points
	LossOfDiffSocketsScore int = 5 // if in different sockets, lose 5 points
)

type OptimalFlextopo struct {
	BestEffortTopologyAlignment bool
}

// type check
var _ framework.CandidatesSortingPlugin = &OptimalFlextopo{}

func NewOptimalFlextopo(_ runtime.Object, _ handle.PodFrameworkHandle) (framework.Plugin, error) {
	// TODO: get args from config
	return &OptimalFlextopo{
		BestEffortTopologyAlignment: true,
	}, nil
}

func (o *OptimalFlextopo) Name() string {
	return OptimalFlextopoName
}

func (o *OptimalFlextopo) Compare(c1, c2 *framework.Candidate) int {
	prioritySum1 := getPrioritySum(c1)
	prioritySum2 := getPrioritySum(c2)
	alignmentScore1 := GetAlignmentScore(c1)
	alignmentScore2 := GetAlignmentScore(c2)
	score1 := getScore(prioritySum1, alignmentScore1)
	score2 := getScore(prioritySum2, alignmentScore2)

	// according to the selection policy of Godel, Candidate[0] is the best;
	// thus, we should put the higher score at the left side.
	if score1 > score2 {
		return 1
	} else if score1 < score2 {
		return -1
	} else {
		return 0
	}
}

func GetAlignmentScore(c *framework.Candidate) int {
	podNames := []string{}
	for _, pod := range c.Victims.Pods {
		podNames = append(podNames, pod.Name)
	}
	podTopologyInfo := ftopoutil.GetPodNUMAAndSockets(c.FlexTopo, podNames)
	tScore := MaxTopologyScore
	totalNumaSet := sets.NewString()
	totalSocketSet := sets.NewString()
	for _, podName := range podNames {
		podTopology := podTopologyInfo[podName]
		totalNumaSet.Insert(podTopology["numas"]...)
		totalSocketSet.Insert(podTopology["sockets"]...)
	}
	if totalNumaSet.Len() > 1 {
		tScore -= LossOfDiffNumasScore
	}
	if totalSocketSet.Len() > 1 {
		tScore -= LossOfDiffSocketsScore
	}

	return tScore
}

// This code is copied from `MinPrioritySum` policy. See:
// `.../sorting/priority/min_priority_sum.go`.
func getPrioritySum(c *framework.Candidate) int64 {
	var sumPriorities int64 = 0
	for _, pod := range c.Victims.Pods {
		// We add MaxInt32+1 to all priorities to make all of them >= 0. This is
		// needed so that a node with a few pods with negative priority is not
		// picked over a node with a smaller number of pods with the same negative
		// priority (and similar scenarios).
		sumPriorities += int64(podutil.GetPodPriority(pod)) + int64(math.MaxInt32+1)
	}
	return sumPriorities
}

func getScore(prioritySum int64, alignmentScore int) float64 {
	if prioritySum == 0 {
		return (1 - PriorityTopologyAlignmentAlpha) * float64(alignmentScore)
	}
	p := 1 / float64(prioritySum)
	t := float64(alignmentScore)
	return PriorityTopologyAlignmentAlpha*p + (1-PriorityTopologyAlignmentAlpha)*t
}
