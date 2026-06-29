package types

import (
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Application defines the structure for an ArgoCD application entry.
type Application struct {
	Name         string                 `yaml:"name"`
	Namespace    string                 `yaml:"namespace"`
	Project      string                 `yaml:"project,omitempty"`
	Path         string                 `yaml:"path,omitempty"`
	Chart        string                 `yaml:"chart,omitempty"`
	ChartVersion string                 `yaml:"chartVersion,omitempty"`
	OtherFields  map[string]interface{} `yaml:",inline"`
}

// Subscription defines the structure for an Operator subscription.
type Subscription struct {
	Name        string                 `yaml:"name"`
	Namespace   string                 `yaml:"namespace"`
	Channel     string                 `yaml:"channel,omitempty"`
	Source      string                 `yaml:"source,omitempty"`
	OtherFields map[string]interface{} `yaml:",inline"`
}

// ClusterGroup holds the detailed configuration for the cluster group.
type ClusterGroup struct {
	Name          string                  `yaml:"name"`
	Namespaces    map[string]interface{}  `yaml:"namespaces"`
	Projects      []string                `yaml:"projects,omitempty"`
	Subscriptions map[string]Subscription `yaml:"subscriptions"`
	Applications  map[string]Application  `yaml:"applications"`
	OtherFields   map[string]interface{}  `yaml:",inline"`
}

// MarshalYAML implements the yaml.Marshaler interface for ClusterGroup.
// It produces compact null representations (key: instead of key: null) for simple namespaces.
func (cg ClusterGroup) MarshalYAML() (interface{}, error) {
	type clusterGroupAlias ClusterGroup
	var doc yaml.Node
	if err := doc.Encode(clusterGroupAlias(cg)); err != nil {
		return nil, err
	}

	node := &doc
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		node = doc.Content[0]
	}

	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content)-1; i += 2 {
			if node.Content[i].Value == "namespaces" && node.Content[i+1].Kind == yaml.MappingNode {
				for j := 1; j < len(node.Content[i+1].Content); j += 2 {
					v := node.Content[i+1].Content[j]
					if v.Kind == yaml.ScalarNode && v.Tag == "!!null" {
						v.Value = ""
					}
				}
				break
			}
		}
	}

	return node, nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for ClusterGroup.
// It handles backward compatibility by converting list-style namespaces to map-style.
func (cg *ClusterGroup) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping node for ClusterGroup, got %d", value.Kind)
	}

	for i := 0; i < len(value.Content)-1; i += 2 {
		keyNode := value.Content[i]
		valNode := value.Content[i+1]

		if keyNode.Value == "namespaces" && valNode.Kind == yaml.SequenceNode {
			newContent := make([]*yaml.Node, 0, len(valNode.Content)*2)
			for _, item := range valNode.Content {
				switch item.Kind {
				case yaml.ScalarNode:
					newContent = append(newContent,
						&yaml.Node{Kind: yaml.ScalarNode, Value: item.Value, Tag: "!!str"},
						&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null"},
					)
				case yaml.MappingNode:
					if len(item.Content) >= 2 {
						newContent = append(newContent, item.Content[0], item.Content[1])
					}
				}
			}
			valNode.Kind = yaml.MappingNode
			valNode.Content = newContent
			valNode.Tag = "!!map"
			break
		}
	}

	type clusterGroupAlias ClusterGroup
	var alias clusterGroupAlias
	if err := value.Decode(&alias); err != nil {
		return err
	}
	*cg = ClusterGroup(alias)
	return nil
}

// ValuesClusterGroup is the top-level struct for the cluster group values file.
type ValuesClusterGroup struct {
	ClusterGroup ClusterGroup           `yaml:"clusterGroup"`
	OtherFields  map[string]interface{} `yaml:",inline"`
}

// NewDefaultValuesClusterGroup creates a default configuration for a cluster group.
// It conditionally includes secrets-related resources based on the useSecrets flag.
func NewDefaultValuesClusterGroup(patternName, clusterGroupName string, chartPaths []string, useSecrets bool) *ValuesClusterGroup {
	namespaces := map[string]interface{}{
		patternName: nil,
	}
	applications := make(map[string]Application)
	subscriptions := make(map[string]Subscription)

	if useSecrets {
		namespaces["vault"] = nil
		namespaces["external-secrets-operator"] = map[string]interface{}{
			"operatorGroup":    true,
			"targetNamespaces": []string{},
		}
		namespaces["external-secrets"] = nil

		subscriptions["eso"] = Subscription{
			Name:      "openshift-external-secrets-operator",
			Namespace: "external-secrets-operator",
			Channel:   "stable-v1",
		}
		applications["vault"] = Application{
			Name:         "vault",
			Namespace:    "vault",
			Chart:        "hashicorp-vault",
			ChartVersion: "0.1.*",
		}
		applications["openshift-external-secrets"] = Application{
			Name:         "openshift-external-secrets",
			Namespace:    "external-secrets",
			Chart:        "openshift-external-secrets",
			ChartVersion: "0.0.*",
		}
	}

	for _, path := range chartPaths {
		chartName := filepath.Base(path)
		app := Application{
			Name:      chartName,
			Namespace: patternName,
			Path:      path,
		}
		applications[chartName] = app
	}

	return &ValuesClusterGroup{
		ClusterGroup: ClusterGroup{
			Name:          clusterGroupName,
			Namespaces:    namespaces,
			Subscriptions: subscriptions,
			Applications:  applications,
			OtherFields:   make(map[string]interface{}),
		},
		OtherFields: make(map[string]interface{}),
	}
}
