package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FlexTopo defines the structure of the CRD
type FlexTopo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlexTopoSpec   `json:"spec,omitempty"`
	Status FlexTopoStatus `json:"status,omitempty"`
}

// FlexTopoSpec defines the content of the topology graph
type FlexTopoSpec struct {
	Nodes []FlexTopoNode `json:"nodes"`
	Edges []FlexTopoEdge `json:"edges"`
}

// FlexTopoNode represents a node in the topology graph
type FlexTopoNode struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Children   []*FlexTopoNode   `json:"children,omitempty"`
}

// FlexTopoEdge represents an edge in the topology graph
type FlexTopoEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
}

// FlexTopoStatus can be used to store status information
type FlexTopoStatus struct {
	// Add status fields as needed
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FlexTopoList is a list of FlexTopo resources
type FlexTopoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FlexTopo `json:"items"`
}
