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

func TestPartitionAddons(t *testing.T) {
	pre, post := partitionAddons(
		[]config.AddonType{config.AddonMultus, config.AddonCertManager, config.AddonMultus},
		config.AddonMultus,
	)
	require.Equal(t, []config.AddonType{config.AddonMultus, config.AddonMultus}, pre)
	require.Equal(t, []config.AddonType{config.AddonCertManager}, post)
}
