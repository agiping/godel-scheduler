package flextopo

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"

	"github.com/kubewharf/godel-scheduler/pkg/binder/framework/handle"
	framework "github.com/kubewharf/godel-scheduler/pkg/framework/api"
)

const (
	Name = "FlexibleTopology"
)

type FlexibleTopology struct {
	GuaranteedTopologyAlignment bool // for binder conflict checking, we only check the guaranteed case
}

var _ framework.CheckConflictsPlugin = &FlexibleTopology{}

func New(_ runtime.Object, handle handle.BinderFrameworkHandle) (framework.Plugin, error) {
	// TODO(Ping Zhang): get args from config
	return &FlexibleTopology{
		GuaranteedTopologyAlignment: true,
	}, nil
}

func (pl *FlexibleTopology) Name() string {
	return Name
}

func (pl *FlexibleTopology) CheckConflicts(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod, nodeInfo framework.NodeInfo) *framework.Status {
	if pl.GuaranteedTopologyAlignment {
		// TODO(Ping Zhang): check topology alignment
		klog.Infof("topology alignment is guaranteed, binder is checking topology alignment for pod %s/%s", pod.Namespace, pod.Name)
	} else {
		klog.Infof("topology alignment is not guaranteed, binder is skipping topology alignment check for pod %s/%s", pod.Namespace, pod.Name)
	}
	return framework.NewStatus(framework.Success, "")
}
