package flextopochecker

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	godelfeatures "github.com/kubewharf/godel-scheduler/pkg/features"
	framework "github.com/kubewharf/godel-scheduler/pkg/framework/api"
	"github.com/kubewharf/godel-scheduler/pkg/scheduler/framework/handle"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
)

const FlexTopoCheckerName string = "FlexTopoChecker"

var _ framework.VictimSearchingPlugin = &FlexTopoChecker{}

type FlexTopoChecker struct {
	GuaranteedTopologyAlignment bool
}

func (ftc *FlexTopoChecker) Name() string {
	return FlexTopoCheckerName
}

func NewFlexTopoChecker(_ runtime.Object, _ handle.PodFrameworkHandle) (framework.Plugin, error) {
	// TODO: get args from config
	return &FlexTopoChecker{
		GuaranteedTopologyAlignment: true,
	}, nil
}

// Placeholder
func (ftc *FlexTopoChecker) VictimSearching(pod *v1.Pod, podInfo *framework.PodInfo, state, _ *framework.CycleState, victimState *framework.VictimState) (framework.Code, string) {
	if !utilfeature.DefaultFeatureGate.Enabled(godelfeatures.FlexibleTopologySupport) {
		// for logging
		return framework.PreemptionSucceed, fmt.Sprintf("featuregate %s is disabled", godelfeatures.FlexibleTopologySupport)
	}
	// TODO(Ping Zhang): revisit the necessity of flextopo checker
	return framework.PreemptionSucceed, ""
}
