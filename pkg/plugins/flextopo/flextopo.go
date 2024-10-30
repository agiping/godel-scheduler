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
	// key: node ID
	Nodes map[string]*FlexNode
}

func NewFlexGraph() *FlexGraph {
	return &FlexGraph{
		Nodes: make(map[string]*FlexNode),
	}
}

func (fg *FlexGraph) String() string {
	var nodeIDs []string
	for id := range fg.Nodes {
		nodeIDs = append(nodeIDs, id)
	}
	return fmt.Sprintf("FlexGraph with nodes: %v", nodeIDs)
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
			if edge.Type == "contains" {
				sourceFlexNode.Children = append(sourceFlexNode.Children, targetFlexNode)
				targetFlexNode.Parents = append(targetFlexNode.Parents, sourceFlexNode)
			}
		}
	}

	return graph
}

// FeasibleFlexibleTopology checks if the node's FlexTopo can satisfy the pod's resource topology requirements.
// This function is only called at the end of filter phase. Thus, it is assumed that all other filters have passed.
// TODO (Ping Zhang): Optimize the performance.
func FeasibleFlexibleTopology(
	pod *v1.Pod,
	resourceType podutil.PodResourceType,
	resourcesRequests map[string]*resource.Quantity,
	nodeInfo framework.NodeInfo,
	podLister listerv1.PodLister,
) *framework.Status {
	// Extract pod's FlexTopo alignment requirements from the pod annotation
	flextopoRequirements := getPodTopologyRequirements(pod)

	if flextopoRequirements == GuaranteedAlignment {
		nodeFlexTopo := nodeInfo.GetFlexTopo()
		flexGraph := BuildFlexGraph(nodeFlexTopo)

		// check for FlexTopo alignment
		canBeAligned, err := checkFlexTopoAlignment(resourcesRequests, flexGraph)
		if err != nil {
			return framework.NewStatus(framework.Error, fmt.Sprintf("failed to check FlexTopo alignment for pod %s on node %s: %v", pod.Name, nodeInfo.GetNodeName(), err))
		}
		if canBeAligned {
			return framework.NewStatus(framework.Success, fmt.Sprintf("FlexTopo alignment can be satisfied for pod %s on node %s", pod.Name, nodeInfo.GetNodeName()))
		} else {
			return framework.NewStatus(framework.Unschedulable, fmt.Sprintf("FlexTopo alignment cannot be satisfied for pod %s on node %s", pod.Name, nodeInfo.GetNodeName()))
		}
	}

	if flextopoRequirements == BestEffortAlignment {
		return framework.NewStatus(framework.Success, "best-effort flextopo alignment is always satisfied during filter phase")
	}

	// For neither guaranteed nor best-effort flextopo alignment requirements, we always return success.
	return framework.NewStatus(framework.Success, "neither guaranteed nor best-effort flextopo alignment requirements is requested")
}

func checkFlexTopoAlignment(
	resourcesRequests map[string]*resource.Quantity,
	flexGraph *FlexGraph,
) (bool, error) {
	// testing
	klog.V(4).InfoS(" ==== checking FlexTopo alignment", "resourcesRequests", resourcesRequests)
	klog.V(4).InfoS(" ==== checking FlexTopo alignment", "node count of flexGraph", len(flexGraph.Nodes))
	klog.V(4).InfoS(" ==== checking FlexTopo alignment", "flexGraph", flexGraph.String())
	return true, nil
}

// getPodTopologyRequirements extracts the FlexTopo alignment requirements from the pod annotations.
func getPodTopologyRequirements(pod *v1.Pod) (flextopoRequirements string) {
	annotations := pod.GetAnnotations()
	if annotations == nil {
		return ""
	}
	return annotations[podutil.FlextopoRequirementAnnotationKey]
}

// // checkTopologyAlignment checks if the requested resources can be allocated with the required topology alignment.
// func checkTopologyAlignment(
// 	resourcesRequests map[string]*resource.Quantity,
// 	topologyGraph *TopologyGraph,
// 	guaranteedAlignment bool,
// ) (bool, error) {
// 	// Extract resource requests
// 	cpuRequest := resourcesRequests["cpu"]
// 	gpuRequest := resourcesRequests["nvidia.com/gpu"]

// 	// For simplicity, assume that resource quantities are in integers
// 	cpuCoresNeeded := int(cpuRequest.Value())
// 	gpusNeeded := 0
// 	if gpuRequest != nil {
// 		gpusNeeded = int(gpuRequest.Value())
// 	}

// 	// Attempt to find aligned resources
// 	if guaranteedAlignment {
// 		// Try to find resources that are fully aligned
// 		aligned, err := findFullyAlignedResources(cpuCoresNeeded, gpusNeeded, topologyGraph)
// 		if err != nil {
// 			return false, err
// 		}
// 		return aligned, nil
// 	} else {
// 		// Best-effort alignment
// 		aligned, err := findBestEffortAlignedResources(cpuCoresNeeded, gpusNeeded, topologyGraph)
// 		if err != nil {
// 			return false, err
// 		}
// 		return aligned, nil
// 	}
// }

// Note: The structs and methods like TopologyGraph, Socket, NUMANode, GetSockets(), GetFreeCPUCoreCount(), etc.,
// need to be defined as per your application's requirements.

// Additionally, implement the methods to parse the FlexTopo CRD into these structures.

// // AddEdge adds an edge to the topology graph
// func (tg *TopologyGraph) AddEdge(edge flextopov1alpha1.FlexTopoEdge) error {
// 	sourceNode, ok := tg.Nodes[edge.Source]
// 	if !ok {
// 		return fmt.Errorf("source node %s not found", edge.Source)
// 	}
// 	targetNode, ok := tg.Nodes[edge.Target]
// 	if !ok {
// 		return fmt.Errorf("target node %s not found", edge.Target)
// 	}

// 	// For this example, we assume the edge type is "contains" or "adjacent"
// 	if edge.Type == "contains" {
// 		// Add targetNode as a child of sourceNode
// 		sourceNode.Children = append(sourceNode.Children, targetNode)
// 		// Add sourceNode as a parent of targetNode
// 		targetNode.Children = append(targetNode.Children, sourceNode)
// 	} else if edge.Type == "adjacent" {
// 		// For "adjacent" type, we might handle it differently if needed
// 		// For now, we can store adjacency information if required
// 		tg.Edges[sourceNode.ID] = append(tg.Edges[sourceNode.ID], targetNode)
// 	}
// 	// You can handle other edge types accordingly
// 	return nil
// }

// // GetSockets returns all socket nodes in the topology graph
// func (tg *TopologyGraph) GetSockets() []*flextopov1alpha1.FlexTopoNode {
// 	sockets := []*flextopov1alpha1.FlexTopoNode{}
// 	for _, node := range tg.Nodes {
// 		if node.Type == "Socket" {
// 			sockets = append(sockets, node)
// 		}
// 	}
// 	return sockets
// }

// // GetNUMANodes returns all NUMA node nodes in the topology graph
// func (tg *TopologyGraph) GetNUMANodes() []*flextopov1alpha1.FlexTopoNode {
// 	numaNodes := []*flextopov1alpha1.FlexTopoNode{}
// 	for _, node := range tg.Nodes {
// 		if node.Type == "NUMANode" {
// 			numaNodes = append(numaNodes, node)
// 		}
// 	}
// 	return numaNodes
// }

// // GetFreeCPUCoreCount returns the number of free CPU cores under this node
// func (tn *flextopov1alpha1.FlexTopoNode) GetFreeCPUCoreCount() int {
// 	count := 0
// 	if tn.Type == "CPUCore" && tn.Attributes["status"] != "used" {
// 		count++
// 	}
// 	for _, child := range tn.Children {
// 		count += child.GetFreeCPUCoreCount()
// 	}
// 	return count
// }

// // GetFreeGPUCount returns the number of free GPUs under this node
// func (tn *flextopov1alpha1.FlexTopoNode) GetFreeGPUCount() int {
// 	count := 0
// 	if tn.Type == "GPU" && tn.Attributes["status"] != "used" {
// 		count++
// 	}
// 	for _, child := range tn.Children {
// 		count += child.GetFreeGPUCount()
// 	}
// 	return count
// }

// // GetFreeCPUCores returns a list of free CPU core nodes under this node
// func (tn *flextopov1alpha1.FlexTopoNode) GetFreeCPUCores() []*flextopov1alpha1.FlexTopoNode {
// 	cores := []*flextopov1alpha1.FlexTopoNode{}
// 	if tn.Type == "CPUCore" && tn.Attributes["status"] != "used" {
// 		cores = append(cores, tn)
// 	}
// 	for _, child := range tn.Children {
// 		cores = append(cores, child.GetFreeCPUCores()...)
// 	}
// 	return cores
// }

// // GetFreeGPUs returns a list of free GPU nodes under this node
// func (tn *flextopov1alpha1.FlexTopoNode) GetFreeGPUs() []*flextopov1alpha1.FlexTopoNode {
// 	gpus := []*flextopov1alpha1.FlexTopoNode{}
// 	if tn.Type == "GPU" && tn.Attributes["status"] != "used" {
// 		gpus = append(gpus, tn)
// 	}
// 	for _, child := range tn.Children {
// 		gpus = append(gpus, child.GetFreeGPUs()...)
// 	}
// 	return gpus
// }

// Example methods to find resources aligned at different topology levels

// // FindAlignedResourcesAtLevel attempts to find resources aligned at the specified topology level
// func (tg *TopologyGraph) FindAlignedResourcesAtLevel(level string, cpuCoresNeeded int, gpusNeeded int) (bool, error) {
// 	switch level {
// 	case "Socket":
// 		for _, socket := range tg.GetSockets() {
// 			if socket.GetFreeCPUCoreCount() >= cpuCoresNeeded && socket.GetFreeGPUCount() >= gpusNeeded {
// 				return true, nil
// 			}
// 		}
// 	case "NUMANode":
// 		for _, numaNode := range tg.GetNUMANodes() {
// 			if numaNode.GetFreeCPUCoreCount() >= cpuCoresNeeded && numaNode.GetFreeGPUCount() >= gpusNeeded {
// 				return true, nil
// 			}
// 		}
// 	// Add other levels like CoreGroup if needed
// 	default:
// 		return false, fmt.Errorf("unknown topology level: %s", level)
// 	}
// 	return false, nil
// }

// Implementing the required methods in the main function

// // findFullyAlignedResources attempts to find resources that are fully aligned on the same socket/NUMA node/core group/GPU.
// func findFullyAlignedResources(
// 	cpuCoresNeeded int,
// 	gpusNeeded int,
// 	topologyGraph *TopologyGraph,
// ) (bool, error) {
// 	// First, try to align at the Socket level
// 	aligned, err := topologyGraph.FindAlignedResourcesAtLevel("Socket", cpuCoresNeeded, gpusNeeded)
// 	if err != nil {
// 		return false, err
// 	}
// 	if aligned {
// 		return true, nil
// 	}
// 	// If not possible, try to align at the NUMA node level
// 	aligned, err = topologyGraph.FindAlignedResourcesAtLevel("NUMANode", cpuCoresNeeded, gpusNeeded)
// 	if err != nil {
// 		return false, err
// 	}
// 	if aligned {
// 		return true, nil
// 	}
// 	// Add more levels if needed
// 	return false, nil
// }

// // findBestEffortAlignedResources attempts to find resources with the best possible alignment.
// func findBestEffortAlignedResources(
// 	cpuCoresNeeded int,
// 	gpusNeeded int,
// 	topologyGraph *TopologyGraph,
// ) (bool, error) {
// 	// Implement a scoring mechanism to find the best alignment
// 	highestScore := -1
// 	bestNode := ""
// 	// We can define a list of levels to try
// 	levels := []string{"Socket", "NUMANode"}
// 	for _, level := range levels {
// 		nodes := []*flextopov1alpha1.FlexTopoNode{}
// 		switch level {
// 		case "Socket":
// 			nodes = topologyGraph.GetSockets()
// 		case "NUMANode":
// 			nodes = topologyGraph.GetNUMANodes()
// 		default:
// 			continue
// 		}
// 		for _, node := range nodes {
// 			score := 0
// 			freeCores := node.GetFreeCPUCoreCount()
// 			freeGPUs := node.GetFreeGPUCount()
// 			if freeCores >= cpuCoresNeeded {
// 				score += 10 // Full CPU alignment
// 			} else {
// 				score += freeCores // Partial CPU alignment
// 			}
// 			if freeGPUs >= gpusNeeded {
// 				score += 10 // Full GPU alignment
// 			} else {
// 				score += freeGPUs // Partial GPU alignment
// 			}
// 			if score > highestScore {
// 				highestScore = score
// 				bestNode = node.ID
// 			}
// 		}
// 	}
// 	if highestScore <= 0 {
// 		// No suitable alignment found
// 		return false, nil
// 	}
// 	klog.V(4).InfoS("Best effort alignment found on node", "node", bestNode, "score", highestScore)
// 	return true, nil
// }

/*
import (
    "fmt"
    "math"
    "sort"

    v1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    framework "k8s.io/kubernetes/pkg/scheduler/framework"
    "k8s.io/kubernetes/pkg/scheduler/framework/plugins/helper"
    listerv1 "k8s.io/client-go/listers/core/v1"
    podutil "k8s.io/kubernetes/pkg/api/v1/pod"
)

// FeasibleFlexibleTopology checks if the node's topology can satisfy the pod's resource requests.
// It considers the NUMA topology and ensures that the resources are allocated in a way that
// minimizes cross-NUMA communication.
func FeasibleFlexibleTopology(
    pod *v1.Pod,
    resourceType podutil.PodResourceType,
    resourcesRequests map[string]*resource.Quantity,
    nodeInfo *framework.NodeInfo,
    podLister listerv1.PodLister,
) *framework.Status {
    // Retrieve the node's FlexTopo data (topology information)
    nodeFlexTopo := nodeInfo.GetFlexTopo()
    if nodeFlexTopo == nil {
        // If the node doesn't have topology information, consider it infeasible
        return framework.NewStatus(framework.Unschedulable, "Node does not have topology information")
    }

    // Extract resource requests from the pod
    cpuRequest := resourcesRequests[v1.ResourceCPU]
    memoryRequest := resourcesRequests[v1.ResourceMemory]
    gpuRequest := resourcesRequests["nvidia.com/gpu"]

    // Handle cases where resource requests are zero or not specified
    cpuReq := int64(0)
    if cpuRequest != nil {
        cpuReq = cpuRequest.MilliValue()
    }
    memReq := int64(0)
    if memoryRequest != nil {
        memReq = memoryRequest.Value()
    }
    gpuReq := int64(0)
    if gpuRequest != nil {
        gpuReq = gpuRequest.Value()
    }

    // Fetch the NUMA nodes from the node's topology
    numaNodes := getNUMANodes(nodeFlexTopo)
    if len(numaNodes) == 0 {
        // If no NUMA nodes are found, consider the node infeasible
        return framework.NewStatus(framework.Unschedulable, "Node does not have NUMA nodes")
    }

    // Build a map of NUMA node IDs to their available resources
    numaResources := getNUMAResources(nodeFlexTopo, numaNodes, nodeInfo)

    // Check if there is any NUMA node that can satisfy the pod's resource requests
    feasible := isFeasibleOnNUMANodes(cpuReq, memReq, gpuReq, numaResources)
    if !feasible {
        return framework.NewStatus(framework.Unschedulable, "Insufficient resources on any NUMA node")
    }

    return nil // Success, the node is feasible
}

// Helper functions used in FeasibleFlexibleTopology:

// getNUMANodes extracts the list of NUMA node IDs from the FlexTopo data.
func getNUMANodes(flexTopo *FlexTopo) []string {
    var numaNodes []string
    for _, node := range flexTopo.Spec.Nodes {
        if node.Type == "NUMANode" {
            numaNodes = append(numaNodes, node.ID)
        }
    }
    return numaNodes
}

// getNUMAResources builds a map of NUMA node IDs to their available resources.
func getNUMAResources(flexTopo *FlexTopo, numaNodes []string, nodeInfo *framework.NodeInfo) map[string]map[v1.ResourceName]int64 {
    // Initialize the resource map for each NUMA node
    numaResources := make(map[string]map[v1.ResourceName]int64)
    for _, numaID := range numaNodes {
        numaResources[numaID] = map[v1.ResourceName]int64{
            v1.ResourceCPU:    0,
            v1.ResourceMemory: 0,
            "nvidia.com/gpu":  0,
        }
    }

    // Aggregate resources per NUMA node
    for _, node := range flexTopo.Spec.Nodes {
        if node.Type == "CPUCore" || node.Type == "Memory" || node.Type == "GPU" {
            // Find the parent NUMA node
            parentNUMA := findParentNUMANode(node, flexTopo)
            if parentNUMA == "" {
                continue
            }
            switch node.Type {
            case "CPUCore":
                if node.Attributes["status"] == "available" {
                    numaResources[parentNUMA][v1.ResourceCPU] += 1
                }
            case "Memory":
                if capacity, ok := node.Attributes["capacity"].(int64); ok {
                    numaResources[parentNUMA][v1.ResourceMemory] += capacity
                }
            case "GPU":
                if node.Attributes["status"] == "available" {
                    numaResources[parentNUMA]["nvidia.com/gpu"] += 1
                }
            }
        }
    }

    // Subtract allocated resources from nodeInfo
    allocatedResources := nodeInfo.Requested
    for _, numaID := range numaNodes {
        // Subtract allocated CPU and memory (assuming uniform distribution)
        numaResources[numaID][v1.ResourceCPU] -= allocatedResources.MilliCPU / int64(len(numaNodes))
        numaResources[numaID][v1.ResourceMemory] -= allocatedResources.Memory / int64(len(numaNodes))
        // Subtract allocated GPUs if applicable
        // (Additional logic may be needed to accurately track GPU allocation per NUMA node)
    }

    return numaResources
}

// findParentNUMANode finds the parent NUMA node ID of a given node.
func findParentNUMANode(node FlexTopoNode, flexTopo *FlexTopo) string {
    // Assuming that the edges define parent-child relationships
    for _, edge := range flexTopo.Spec.Edges {
        if edge.Target == node.ID && edge.Type == "parent" {
            return edge.Source
        }
    }
    return ""
}

// isFeasibleOnNUMANodes checks if any NUMA node can satisfy the resource requests.
func isFeasibleOnNUMANodes(cpuReq, memReq, gpuReq int64, numaResources map[string]map[v1.ResourceName]int64) bool {
    for _, resources := range numaResources {
        if resources[v1.ResourceCPU] >= cpuReq &&
            resources[v1.ResourceMemory] >= memReq &&
            resources["nvidia.com/gpu"] >= gpuReq {
            return true
        }
    }
    return false
}
*/
