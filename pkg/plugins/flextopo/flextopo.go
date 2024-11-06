package flextopo

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	listerv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"

	flextopov1alpha1 "github.com/agiping/flextopo-api/pkg/apis/flextopo/v1alpha1"
	framework "github.com/kubewharf/godel-scheduler/pkg/framework/api"
	podutil "github.com/kubewharf/godel-scheduler/pkg/util/pod"
)

const (
	BestEffortAlignment string = "best-effort"
	GuaranteedAlignment string = "guaranteed"
	NodeTypeNUMA        string = "NUMANode"
	NodeTypeSocket      string = "Socket"
	NodeTypeCPUCore     string = "CPUCore"
	NodeTypeCoreGroup   string = "CoreGroup"
	NodeTypeGPU         string = "GPU"
	EdgeTypeContains    string = "contains"

	AlignmentLevelNUMA   string = "NUMA"
	AlignmentLevelSocket string = "Socket"
	AlignmentLevelFailed string = "Failed"
)

// FlexNode represents a node in the FlexTopo graph
type FlexNode struct {
	ID         string
	Type       string
	Attributes map[string]string
	Parents    []*FlexNode
	Children   []*FlexNode
}

// FlexGraph represents the FlexTopo graph
type FlexGraph struct {
	Nodes                 map[string]*FlexNode // key: node ID
	FreeCPUCoresByNUMAs   map[string]int       // key: NUMA ID
	FreeCPUCoresBySockets map[string]int       // key: Socket ID
}

func NewFlexGraph() *FlexGraph {
	return &FlexGraph{
		Nodes:                 make(map[string]*FlexNode),
		FreeCPUCoresByNUMAs:   make(map[string]int),
		FreeCPUCoresBySockets: make(map[string]int),
	}
}

// BuildFlexGraph builds a FlexGraph from a FlexTopo CRD
func BuildFlexGraph(flexTopo *flextopov1alpha1.FlexTopo) *FlexGraph {
	graph := NewFlexGraph()

	// build node map
	for _, node := range flexTopo.Spec.Nodes {
		flexNode := &FlexNode{
			ID:         node.ID,
			Type:       node.Type,
			Attributes: node.Attributes,
			Parents:    []*FlexNode{},
			Children:   []*FlexNode{},
		}
		graph.Nodes[node.ID] = flexNode
	}

	// build node connections (parent-child)
	for _, edge := range flexTopo.Spec.Edges {
		sourceFlexNode := graph.Nodes[edge.Source]
		targetFlexNode := graph.Nodes[edge.Target]
		if sourceFlexNode != nil && targetFlexNode != nil {
			if edge.Type == EdgeTypeContains {
				sourceFlexNode.Children = append(sourceFlexNode.Children, targetFlexNode)
				targetFlexNode.Parents = append(targetFlexNode.Parents, sourceFlexNode)
			}
		}
	}

	// compute free CPU cores for each NUMA node and socket
	graph.ComputeFreeCPUCoresByNUMAsAndSockets()

	return graph
}

func (fg *FlexGraph) String() string {
	var nodeIDs []string
	for id := range fg.Nodes {
		nodeIDs = append(nodeIDs, id)
	}
	return fmt.Sprintf("FlexGraph with nodes: %v", nodeIDs)
}

func (fg *FlexGraph) ComputeFreeCPUCoresByNUMAsAndSockets() {
	numaNodes, socketNodes := fg.GetFlexNodesByType()
	// compute free CPU cores for each NUMA node
	for _, numaID := range numaNodes {
		fg.FreeCPUCoresByNUMAs[numaID] = 0
		for _, coreGroupID := range fg.Nodes[numaID].Children {
			if coreGroupID.Type == NodeTypeCoreGroup {
				for _, coreID := range coreGroupID.Children {
					if coreID.Type == NodeTypeCPUCore {
						if coreID.Attributes["status"] == "free" {
							fg.FreeCPUCoresByNUMAs[numaID]++
						}
					}
				}
			}
		}
	}

	// compute free CPU cores for each socket
	for _, socketID := range socketNodes {
		fg.FreeCPUCoresBySockets[socketID] = 0
		for _, numaID := range fg.Nodes[socketID].Children {
			fg.FreeCPUCoresBySockets[socketID] += fg.FreeCPUCoresByNUMAs[numaID.ID]
		}
	}

	// for validation
	klog.V(4).InfoS("Free CPU cores by NUMAs", "numa_free_cores", fg.FreeCPUCoresByNUMAs)
	klog.V(4).InfoS("Free CPU cores by sockets", "socket_free_cores", fg.FreeCPUCoresBySockets)
}

// GetFlexNodesByType extracts the list of node IDs of the type NUMA and Socket from the FlexGraph.
func (fg *FlexGraph) GetFlexNodesByType() (numaNodes []string, sockets []string) {
	for _, node := range fg.Nodes {
		if node.Type == NodeTypeNUMA {
			numaNodes = append(numaNodes, node.ID)
		}
		if node.Type == NodeTypeSocket {
			sockets = append(sockets, node.ID)
		}
	}
	return numaNodes, sockets
}

// MostLocalizedFreeCPUCoresByNUMAs returns the NUMA node with the most localized free CPU cores.
func (fg *FlexGraph) MostLocalizedFreeCPUCoresByNUMAs() (string, int) {
	mostLocalizedFreeCPUCores := 0
	mostLocalizedFreeNUMANode := ""
	for numaID, freeCPUCores := range fg.FreeCPUCoresByNUMAs {
		if freeCPUCores > mostLocalizedFreeCPUCores {
			mostLocalizedFreeCPUCores = freeCPUCores
			mostLocalizedFreeNUMANode = numaID
		}
	}
	return mostLocalizedFreeNUMANode, mostLocalizedFreeCPUCores
}

// MostHostedFreeCPUCoresBySockets returns the socket with the most hosted free CPU cores.
func (fg *FlexGraph) MostHostedFreeCPUCoresBySockets() (string, int) {
	mostHostedFreeCPUCores := 0
	mostHostedFreeSocket := ""
	for socketID, freeCPUCores := range fg.FreeCPUCoresBySockets {
		if freeCPUCores > mostHostedFreeCPUCores {
			mostHostedFreeCPUCores = freeCPUCores
			mostHostedFreeSocket = socketID
		}
	}
	return mostHostedFreeSocket, mostHostedFreeCPUCores
}

// FeasibleFlexibleTopology checks if the node's FlexTopo can satisfy the pod's resource topology requirements.
// This function is only called at the end of filter phase. Thus, it is assumed that all other filters have passed.
// ATTENTION: the scheduler only checks the feasibility of alignment, it does not do the actual alignment.
// Actual alignment is done by the kubelet during pod admission.
// TODO (Ping Zhang): Optimize the performance.
func FeasibleFlexibleTopology(
	pod *v1.Pod,
	resourceType podutil.PodResourceType,
	resourcesRequests map[string]*resource.Quantity,
	nodeInfo framework.NodeInfo,
	podLister listerv1.PodLister,
) *framework.Status {
	// Extract pod's FlexTopo alignment requirements from the pod annotation
	flextopoRequirements := podutil.GetPodTopologyRequirements(pod)

	if flextopoRequirements == GuaranteedAlignment {
		nodeFlexTopo := nodeInfo.GetFlexTopo()
		// TODO(Ping Zhang):
		// 1. check if victims are already removed from the nodeInfo during preemption.
		// 2. check if we need to remove victims from the nodeFlexTopo for preemption.
		// 3. we have already validated that current code is able to handle the normal scheduling cycle.
		flexGraph := BuildFlexGraph(nodeFlexTopo)

		// check for FlexTopo alignment
		canBeAligned, level := checkFlexTopoAlignment(resourcesRequests, flexGraph)
		if canBeAligned {
			return framework.NewStatus(framework.Success, fmt.Sprintf("FlexTopo alignment can be satisfied for pod %s on node %s, at %s level", pod.Name, nodeInfo.GetNodeName(), level))
		} else {
			return framework.NewStatus(framework.Unschedulable, fmt.Sprintf("FlexTopo alignment cannot be satisfied for pod %s on node %s", pod.Name, nodeInfo.GetNodeName()))
		}
	}

	if flextopoRequirements == BestEffortAlignment {
		return framework.NewStatus(framework.Success, "best-effort flextopo alignment is always satisfied during filter phase")
	}

	// For neither guaranteed nor best-effort flextopo alignment requirements, we always return success.
	// This would not be happen, but we keep it for robustness.
	return framework.NewStatus(framework.Success, "neither guaranteed nor best-effort flextopo alignment requirements is requested")
}

func checkFlexTopoAlignment(resourcesRequests map[string]*resource.Quantity, flexGraph *FlexGraph) (bool, string) {
	cpuRequest := resourcesRequests["cpu"]
	cpuCoresNeeded := int(cpuRequest.Value())

	// step1: try to align on the same NUMA node
	canBeAlignedByNUMA := flexGraph.TryAlignOnSameNUMANode(cpuCoresNeeded)

	if canBeAlignedByNUMA {
		return true, AlignmentLevelNUMA
	}
	// step2: try to align on the same socket
	canBeAlignedBySocket := flexGraph.TryAlignOnSameSocket(cpuCoresNeeded)
	if canBeAlignedBySocket {
		return true, AlignmentLevelSocket
	}

	// step3: no feasible alignment
	return false, AlignmentLevelFailed
}

func (fg *FlexGraph) TryAlignOnSameNUMANode(cpuCoresNeeded int) bool {
	// If the largest number of free CPU cores under any NUMA node is less than the number of CPU cores needed,
	// it means that none of the single NUMA node can satisfy the flextopo alignment requirement.
	mostLocalizedFreeNUMANode, mostLocalizedFreeCPUCores := fg.MostLocalizedFreeCPUCoresByNUMAs()
	if mostLocalizedFreeCPUCores < cpuCoresNeeded {
		return false
	}
	// Otherwise, the pod can, at least, be aligned on the NUMA node with the most localized free CPU cores.
	klog.V(4).InfoS("FlexTopo alignment can be satisfied on NUMA node", "node", mostLocalizedFreeNUMANode)
	return true
}

func (fg *FlexGraph) TryAlignOnSameSocket(cpuCoresNeeded int) bool {
	// If the largest number of free CPU cores under any socket is less than the number of CPU cores needed,
	// it means that none of the single socket can satisfy the flextopo alignment requirement.
	// Socket level alignment is the last resort for flextopo alignment.
	mostHostedFreeSocket, mostHostedFreeCPUCores := fg.MostHostedFreeCPUCoresBySockets()
	if mostHostedFreeCPUCores < cpuCoresNeeded {
		return false
	}
	// Otherwise, the pod can, at least, be aligned on the socket with the most hosted free CPU cores.
	klog.V(4).InfoS("FlexTopo alignment can be satisfied on socket", "socket:", mostHostedFreeSocket)
	return true
}
