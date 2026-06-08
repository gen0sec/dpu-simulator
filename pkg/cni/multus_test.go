package cni

import (
	"testing"

	"github.com/ovn-kubernetes/dpu-simulator/pkg/config"
)

func TestEnsureMultusReadySkipsWithoutMultusAddon(t *testing.T) {
	m := &CNIManager{
		config: &config.Config{
			Kubernetes: config.KubernetesConfig{
				Clusters: []config.ClusterConfig{{Name: "host", CNI: config.CNIOVNKubernetes}},
			},
		},
	}

	if err := m.ensureMultusReady("host"); err != nil {
		t.Fatalf("ensureMultusReady() without multus addon: %v", err)
	}
}

func TestEnsureMultusReadySkipsValuesOnlyOVNK(t *testing.T) {
	m := &CNIManager{
		config: &config.Config{
			OVNKubernetesMode: config.OVNKubernetesModeValuesOnly,
			Kubernetes: config.KubernetesConfig{
				Clusters: []config.ClusterConfig{
					{Name: "host", CNI: config.CNIOVNKubernetes, Addons: []config.AddonType{config.AddonMultus}},
				},
			},
		},
	}

	if err := m.ensureMultusReady("host"); err != nil {
		t.Fatalf("ensureMultusReady() values-only OVN-K: %v", err)
	}
}
