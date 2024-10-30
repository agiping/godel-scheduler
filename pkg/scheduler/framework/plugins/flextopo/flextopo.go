package flextopo

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	listerv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"

	godelfeatures "github.com/kubewharf/godel-scheduler/pkg/features"
	framework "github.com/kubewharf/godel-scheduler/pkg/framework/api"
	flextopocrd "github.com/kubewharf/godel-scheduler/pkg/plugins/flextopo"
	"github.com/kubewharf/godel-scheduler/pkg/scheduler/framework/handle"
	podutil "github.com/kubewharf/godel-scheduler/pkg/util/pod"
)

const Name = "FlexibleTopology"

type FlexibleTopology struct {
	podLister listerv1.PodLister
}

var (
	_ framework.FilterPlugin = &FlexibleTopology{}
	_ framework.ScorePlugin  = &FlexibleTopology{}
)

func New(_ runtime.Object, handle handle.PodFrameworkHandle) (framework.Plugin, error) {
	klog.V(4).InfoS("====== Creating flextopo plugin ======")
	informerFactory := handle.SharedInformerFactory()
	podLister := informerFactory.Core().V1().Pods().Lister()
	return &FlexibleTopology{
		podLister: podLister,
	}, nil
}

func (flextopo *FlexibleTopology) Name() string {
	return Name
}

func (flextopo *FlexibleTopology) PreFilter(_ context.Context, cycleState *framework.CycleState, pod *v1.Pod) *framework.Status {
	if !utilfeature.DefaultFeatureGate.Enabled(godelfeatures.FlexibleTopologySupport) {
		return framework.NewStatus(framework.Error, fmt.Sprintf("featuregate %s is disabled", godelfeatures.FlexibleTopologySupport))
	}
	// Get all resources requests
	resourcesRequests := podutil.GetPodRequests(pod)
	// TODO(Ping Zhang): do necessary pre-filtering here
	klog.V(4).InfoS("====== Running pre-filter function of flextopo plugin ======", "resourcesRequests", resourcesRequests)
	return nil
}

func (flextopo *FlexibleTopology) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

func (flextopo *FlexibleTopology) Filter(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod, nodeInfo framework.NodeInfo) *framework.Status {
	// testing purpose
	klog.V(4).InfoS("====== Running filter function of flextopo plugin ======")
	if !utilfeature.DefaultFeatureGate.Enabled(godelfeatures.FlexibleTopologySupport) {
		return framework.NewStatus(framework.Error, fmt.Sprintf("featuregate %s is disabled", godelfeatures.FlexibleTopologySupport))
	}
	// TODO(Ping Zhang): do necessary filtering here
	resourceType, err := framework.GetPodResourceType(cycleState)
	if err != nil {
		return framework.NewStatus(framework.Error, "failed to get resource type from state")
	}
	resourcesRequests := podutil.GetPodRequests(pod)
	return flextopocrd.FeasibleFlexibleTopology(pod, resourceType, resourcesRequests, nodeInfo, flextopo.podLister)
}

func (flextopo *FlexibleTopology) PreScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []framework.NodeInfo) *framework.Status {
	// placeholder
	klog.V(4).InfoS("====== Running pre-score function of flextopo plugin ======")
	return nil
}

// TODO(Ping Zhang): implement score function
func (flextopo *FlexibleTopology) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	// placeholder
	klog.V(4).InfoS("====== Running score function of flextopo plugin ======")
	return 0, nil
}

// TODO(Ping Zhang): implement score extensions
func (flextopo *FlexibleTopology) ScoreExtensions() framework.ScoreExtensions {
	// placeholder
	klog.V(4).InfoS("====== Running score extensions of flextopo plugin ======")
	return nil
}
