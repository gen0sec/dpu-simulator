package cni

import (
	"testing"

	"github.com/ovn-kubernetes/dpu-simulator/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestResolveAddonInstallOrderPassthroughWithoutWhereaboutsAddon(t *testing.T) {
	addons := []config.AddonType{config.AddonMultus, config.AddonCertManager}
	ordered := resolveAddonInstallOrder(addons)
	require.Equal(t, addons, ordered)
}

func TestResolveAddonInstallOrderDoesNotDuplicateWhereabouts(t *testing.T) {
	addons := []config.AddonType{config.AddonWhereabouts, config.AddonMultus, config.AddonCertManager}
	ordered := resolveAddonInstallOrder(addons)
	require.Equal(t, addons, ordered)
}

func TestPartitionPreCNIAddons(t *testing.T) {
	ordered := []config.AddonType{config.AddonWhereabouts, config.AddonMultus, config.AddonCertManager}
	pre, post := partitionPreCNIAddons(ordered)
	require.Equal(t, []config.AddonType{config.AddonWhereabouts, config.AddonMultus}, pre)
	require.Equal(t, []config.AddonType{config.AddonCertManager}, post)
}

func TestClusterHasAddon(t *testing.T) {
	m := &CNIManager{
		config: &config.Config{
			Kubernetes: config.KubernetesConfig{
				Clusters: []config.ClusterConfig{
					{Name: "host", Addons: []config.AddonType{config.AddonMultus, config.AddonWhereabouts}},
					{Name: "plain"},
				},
			},
		},
	}

	require.True(t, m.clusterHasAddon("host", config.AddonMultus))
	require.False(t, m.clusterHasAddon("host", config.AddonCertManager))
	require.False(t, m.clusterHasAddon("plain", config.AddonMultus))
	require.False(t, m.clusterHasAddon("missing", config.AddonMultus))
}
