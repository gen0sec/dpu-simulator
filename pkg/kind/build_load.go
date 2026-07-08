package kind

import (
	"context"
	"fmt"

	"github.com/ovn-kubernetes/dpu-simulator/pkg/cni"
	"github.com/ovn-kubernetes/dpu-simulator/pkg/config"
	"github.com/ovn-kubernetes/dpu-simulator/pkg/containerengine"
	"github.com/ovn-kubernetes/dpu-simulator/pkg/deviceplugin"
	"github.com/ovn-kubernetes/dpu-simulator/pkg/log"
	"github.com/ovn-kubernetes/dpu-simulator/pkg/platform"
)

// BuildAndLoadImagesFromRegistryConfig builds images declared under registry.containers
// when the registry is disabled (Kind mode). Images are loaded into Kind so nodes can run
// them; pod specs must use the matching localhost/… references from config.KindNodeLocalImageRef.
func (m *KindManager) BuildAndLoadImagesFromRegistryConfig(cmdExec platform.CommandExecutor, engine containerengine.Engine) error {
	cfg := m.config
	if !cfg.IsRegistryImageBuildOnly() || !cfg.IsKindMode() {
		return nil
	}

	log.Info("\n=== Building registry container images (registry disabled; loading into Kind) ===")
	build := cni.BuildCNIImageWithRuntime(cfg, cmdExec, engine)

	for _, container := range cfg.Registry.Containers {
		localImage, err := build(container)
		if err != nil {
			return fmt.Errorf("build registry container %q: %w", container.Name, err)
		}

		switch config.CNIType(container.CNI) {
		case config.CNIOVNKubernetes:
			kindRef, err := m.prepareBuiltImageForKindLoad(context.Background(), cmdExec, engine, localImage)
			if err != nil {
				return fmt.Errorf("prepare registry container %q image for Kind: %w", container.Name, err)
			}
			var ovnkClusters []string
			for _, cl := range cfg.Kubernetes.Clusters {
				if cfg.ClusterNeedsOVNKubernetesImage(cl.Name) {
					ovnkClusters = append(ovnkClusters, cl.Name)
				}
			}
			if err := m.KindLoadImageIntoClusters(cmdExec, kindRef, ovnkClusters); err != nil {
				return fmt.Errorf("kind load %q into clusters %v: %w", kindRef, ovnkClusters, err)
			}
		default:
			return fmt.Errorf("unsupported CNI type for registry container build: %q", container.CNI)
		}
	}

	if cfg.IsOffloadDPU() {
		if err := deviceplugin.BuildDevicePluginImage(cmdExec, engine); err != nil {
			return fmt.Errorf("build device plugin image: %w", err)
		}
		hostCluster := cfg.GetDPUHostClusterName()
		if hostCluster != "" {
			kindRef, err := m.prepareBuiltImageForKindLoad(context.Background(), cmdExec, engine, deviceplugin.DevicePluginImage)
			if err != nil {
				return fmt.Errorf("prepare device plugin image for Kind: %w", err)
			}
			if err := m.KindLoadImage(cmdExec, hostCluster, kindRef); err != nil {
				return fmt.Errorf("kind load device plugin into cluster %q: %w", hostCluster, err)
			}
			log.Info("✓ Device plugin image loaded into cluster %s", hostCluster)
		}
	}

	log.Info("✓ Registry image builds complete; images loaded into Kind")
	return nil
}

// prepareBuiltImageForKindLoad ensures builtRef exists under config.KindNodeLocalImageRef(builtRef) in
// m.containerBin's image store so KindLoadImage can save it for kind load image-archive.
func (m *KindManager) prepareBuiltImageForKindLoad(
	ctx context.Context,
	cmdExec platform.CommandExecutor,
	engine containerengine.Engine,
	builtRef string,
) (string, error) {
	nodeLocal := config.KindNodeLocalImageRef(builtRef)
	if nodeLocal == "" {
		return "", fmt.Errorf("empty Kind node-local image ref derived from %q", builtRef)
	}

	engineBin := string(engine.Name())

	if engineBin != m.containerBin {
		if nodeLocal != builtRef {
			if err := engine.Tag(ctx, builtRef, nodeLocal); err != nil {
				return "", fmt.Errorf(
					"retag built image %q to Kind node-local ref %q with build engine %s (Kind uses %s): %w",
					builtRef, nodeLocal, engineBin, m.containerBin, err,
				)
			}
		}
		if err := containerengine.TransferImageBetweenRuntimes(cmdExec, engineBin, m.containerBin, nodeLocal); err != nil {
			return "", fmt.Errorf(
				"copy image into %s after build with %s (engine/runtime mismatch): %w",
				m.containerBin, engineBin, err,
			)
		}
		if err := containerengine.ImagePresentInRuntime(cmdExec, m.containerBin, nodeLocal); err != nil {
			return "", fmt.Errorf("%w (required before kind load image-archive)", err)
		}
		return nodeLocal, nil
	}

	if nodeLocal != builtRef {
		if err := engine.Tag(ctx, builtRef, nodeLocal); err != nil {
			return "", fmt.Errorf(
				"retag built image %q to Kind node-local ref %q with %s: %w",
				builtRef, nodeLocal, engineBin, err,
			)
		}
	}
	if err := containerengine.ImagePresentInRuntime(cmdExec, m.containerBin, nodeLocal); err != nil {
		return "", fmt.Errorf("%w (required before kind load image-archive)", err)
	}
	return nodeLocal, nil
}
