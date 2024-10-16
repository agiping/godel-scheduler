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
	"k8s.io/apimachinery/pkg/runtime"

	framework "github.com/kubewharf/godel-scheduler/pkg/framework/api"
	"github.com/kubewharf/godel-scheduler/pkg/scheduler/framework/handle"
)

const OptimalFlextopoName = "OptimalFlextopo"

type OptimalFlextopo struct{}

// type check
var _ framework.CandidatesSortingPlugin = &OptimalFlextopo{}

func NewOptimalFlextopo(_ runtime.Object, _ handle.PodFrameworkHandle) (framework.Plugin, error) {
	return &OptimalFlextopo{}, nil
}

func (o *OptimalFlextopo) Name() string {
	return OptimalFlextopoName
}

func (o *OptimalFlextopo) Compare(c1, c2 *framework.Candidate) int {
	// TODO: implement the comparison logic
	return 0
}
