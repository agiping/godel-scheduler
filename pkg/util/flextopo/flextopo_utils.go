package flextopo

import (
	"k8s.io/klog/v2"

	flextopov1alpha1 "github.com/agiping/flextopo-api/pkg/apis/flextopo/v1alpha1"
)

// PodTopologyInfo is a map of pod names to their NUMA node IDs and Socket IDs.
// key: pod name
// value: a map of node type to node IDs
// e.g.,
//
//		{
//		  "pod1": {
//		    "numas": ["numa0", "numa1"],
//		    "sockets": ["socket0",]
//		  }
//	   ...
//		}
type PodTopologyInfo map[string]map[string][]string

// GetPodNUMAAndSockets extracts the NUMA nodes and Socket IDs for a list of pod names.
func GetPodNUMAAndSockets(flexTopo *flextopov1alpha1.FlexTopo, podNames []string) PodTopologyInfo {
	// Build maps for quick lookups
	if flexTopo == nil {
		klog.Errorf("FlexTopo is nil, cannot get pod NUMA and socket info")
		return nil
	}
	nodeMap := make(map[string]*flextopov1alpha1.FlexTopoNode)
	for i := range flexTopo.Spec.Nodes {
		node := &flexTopo.Spec.Nodes[i]
		nodeMap[node.ID] = node
	}

	// Build child-to-parent map
	childToParent := make(map[string]string)
	for _, edge := range flexTopo.Spec.Edges {
		if edge.Type == "contains" {
			childToParent[edge.Target] = edge.Source
		}
	}

	result := make(PodTopologyInfo)

	for _, podName := range podNames {
		// Find all CPU cores used by the pod
		cpuCores := []string{}
		for _, node := range flexTopo.Spec.Nodes {
			if node.Type == "CPUCore" && node.Attributes["usedBy"] == podName {
				cpuCores = append(cpuCores, node.ID)
			}
		}

		// Collect NUMA nodes and Sockets
		numaNodes := make(map[string]bool)
		sockets := make(map[string]bool)

		for _, coreID := range cpuCores {
			currentID := coreID
			for {
				parentID, exists := childToParent[currentID]
				if !exists {
					break
				}
				parentNode, exists := nodeMap[parentID]
				if !exists {
					break
				}

				if parentNode.Type == "NUMANode" {
					numaNodes[parentNode.ID] = true
				}
				if parentNode.Type == "Socket" {
					sockets[parentNode.ID] = true
				}

				currentID = parentID
			}
		}

		// Convert maps to slices
		numaList := []string{}
		for numaID := range numaNodes {
			numaList = append(numaList, numaID)
		}

		socketList := []string{}
		for socketID := range sockets {
			socketList = append(socketList, socketID)
		}

		// Store in result map
		result[podName] = map[string][]string{
			"numas":   numaList,
			"sockets": socketList,
		}
	}

	return result
}

// TODO(Ping Zhang): Provide a generic function to parse the hardware topology
// We can cache the parsed topology of each node in the framework handle for reuse
func ParseHardwareTopology(flexTopo *flextopov1alpha1.FlexTopo) {
	klog.Infof("flexTopo: %v", flexTopo)
}
