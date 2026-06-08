package deviceplugin

import (
	"fmt"
	"testing"

	"github.com/ovn-kubernetes/dpu-simulator/lib/dpusim"
	"github.com/ovn-kubernetes/dpu-simulator/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildResourcePools verifies mgmt and pod VF pools partition host data
// interfaces correctly for several mgmt_port_vfs_count values, including
// MatcherDescription output and non-overlapping MatchesIface selectors.
func TestBuildResourcePools(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mgmtCount      int
		mgmtMatches    []string
		mgmtNonMatches []string
		podMatches     []string
		podNonMatches  []string
	}{
		{
			name:           "two mgmt VFs",
			mgmtCount:      2,
			mgmtMatches:    []string{dpusim.HostDataIf(1), dpusim.HostDataIf(2)},
			mgmtNonMatches: []string{dpusim.HostDataIf(0), dpusim.HostDataIf(3), dpusim.DPUDataIf(1)},
			podMatches:     []string{dpusim.HostDataIf(3), dpusim.HostDataIf(10)},
			podNonMatches:  []string{dpusim.HostDataIf(0), dpusim.HostDataIf(1), dpusim.HostDataIf(2)},
		},
		{
			name:           "three mgmt VFs",
			mgmtCount:      3,
			mgmtMatches:    []string{dpusim.HostDataIf(1), dpusim.HostDataIf(2), dpusim.HostDataIf(3)},
			mgmtNonMatches: []string{dpusim.HostDataIf(0), dpusim.HostDataIf(4)},
			podMatches:     []string{dpusim.HostDataIf(4), dpusim.HostDataIf(15)},
			podNonMatches:  []string{dpusim.HostDataIf(0), dpusim.HostDataIf(1), dpusim.HostDataIf(2), dpusim.HostDataIf(3)},
		},
		{
			name:           "single mgmt VF",
			mgmtCount:      1,
			mgmtMatches:    []string{dpusim.HostDataIf(1)},
			mgmtNonMatches: []string{dpusim.HostDataIf(0), dpusim.HostDataIf(2)},
			podMatches:     []string{dpusim.HostDataIf(2), dpusim.HostDataIf(9)},
			podNonMatches:  []string{dpusim.HostDataIf(0), dpusim.HostDataIf(1)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pools, err := BuildResourcePools(tt.mgmtCount)
			require.NoError(t, err)
			assert.Len(t, pools, 2)

			mgmtPool := pools[0]
			podPool := pools[1]
			assert.Equal(t, MgmtVFResourceName, mgmtPool.ResourceName)
			assert.Equal(t, VFResourceName, podPool.ResourceName)

			if tt.mgmtCount == 1 {
				assert.Equal(t, dpusim.HostDataIf(1), mgmtPool.MatcherDescription())
			} else {
				assert.Equal(t, fmt.Sprintf("%s..%s", dpusim.HostDataIf(1), dpusim.HostDataIf(tt.mgmtCount)), mgmtPool.MatcherDescription())
			}
			assert.Equal(t, fmt.Sprintf("%s..", dpusim.HostDataIf(tt.mgmtCount+1)), podPool.MatcherDescription())

			for _, iface := range tt.mgmtMatches {
				assert.True(t, mgmtPool.MatchesIface(iface), "mgmt should match %s", iface)
				assert.False(t, podPool.MatchesIface(iface), "pod should not match %s", iface)
			}
			for _, iface := range tt.mgmtNonMatches {
				assert.False(t, mgmtPool.MatchesIface(iface), "mgmt should not match %s", iface)
			}
			for _, iface := range tt.podMatches {
				assert.True(t, podPool.MatchesIface(iface), "pod should match %s", iface)
				assert.False(t, mgmtPool.MatchesIface(iface), "mgmt should not match %s", iface)
			}
			for _, iface := range tt.podNonMatches {
				assert.False(t, podPool.MatchesIface(iface), "pod should not match %s", iface)
			}
		})
	}
}

// TestGatewayInterfaceExcludedFromPools verifies eth0-0 is never advertised by
// either resource pool regardless of mgmt_port_vfs_count.
func TestGatewayInterfaceExcludedFromPools(t *testing.T) {
	t.Parallel()

	gateway := dpusim.HostGatewayInterface
	for _, mgmtCount := range []int{1, 2, 3, config.DefaultMgmtPortVFsCount} {
		pools, err := BuildResourcePools(mgmtCount)
		require.NoError(t, err)
		for _, pool := range pools {
			assert.False(t, pool.MatchesIface(gateway), "pool %s must not match %s", pool.ResourceName, gateway)
		}
	}
}

func TestBuildResourcePoolsInvalidCountErrors(t *testing.T) {
	t.Parallel()

	_, err := BuildResourcePools(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mgmt_port_vfs_count")
}

// TestMgmtPortVFsCountFromEnv checks parsing of MGMT_PORT_VFS_COUNT.
func TestMgmtPortVFsCountFromEnv(t *testing.T) {
	t.Setenv(MgmtPortVFsCountEnvVar, "5")
	count, err := MgmtPortVFsCountFromEnv()
	require.NoError(t, err)
	assert.Equal(t, 5, count)

	t.Setenv(MgmtPortVFsCountEnvVar, "")
	_, err = MgmtPortVFsCountFromEnv()
	require.Error(t, err)
	assert.Contains(t, err.Error(), MgmtPortVFsCountEnvVar)

	t.Setenv(MgmtPortVFsCountEnvVar, "invalid")
	_, err = MgmtPortVFsCountFromEnv()
	require.Error(t, err)
	assert.Contains(t, err.Error(), MgmtPortVFsCountEnvVar)
}
