package generator

import (
	"fmt"

	composetypes "github.com/compose-spec/compose-go/types"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/servicetypes"
)

var (
	maxAdditionalVolumes        int    = 6
	defaultAdditionalVolumeSize string = "5Gi"
	maxAdditionalVolumeSize     string = "4Ti"
)

// convertVolumes handles converting docker compose volumes into lagoon volumes and adds them to build values
func convertVolumes(buildValues *BuildValues, lCompose *composetypes.Project, lComposeVolumes []lagoon.OriginalVolumeOrder, debug bool) error {
	// convert docker-compose volumes to buildvolumes,
	// range over the volumes and add them to build values
	for _, vol := range lComposeVolumes {
		for _, composeVolumeValues := range lCompose.Volumes {
			// check that the volumename from the ordered volumes matches (with the composestack name prefix)
			if lagoon.GetComposeVolumeName(lCompose.Name, vol.Name) == composeVolumeValues.Name {
				// if so, check that the volume returns values correctly
				cVolume, err := composeToVolumeValues(lCompose.Name, composeVolumeValues, debug)
				if err != nil {
					return err
				}
				if cVolume != nil {
					buildValues.Volumes = append(buildValues.Volumes, *cVolume)
				}
				// to prevent too many volumes from being provisioned, some sort of limit should probably be imposed
				// harcdoded now to maxAdditionalVolumes, but could be configurable
				if len(buildValues.Volumes) > maxAdditionalVolumes {
					return fmt.Errorf("unable to provision more than %d volumes for this environment, if you need more please contact your lagoon administrator", maxAdditionalVolumes)
				}
			}
		}
	}
	return nil
}

// composeToVolumeValues handles converting a docker compose volume and the labels into a lagoon volume
func composeToVolumeValues(
	composeName string,
	composeVolumeValues composetypes.VolumeConfig,
	debug bool,
) (*ComposeVolume, error) {
	// if there are no labels, then this is probably not going to end up in Lagoon
	// the lagoonType check will skip to the end and return an empty service definition
	if composeVolumeValues.Labels != nil {
		volumeType := lagoon.CheckDockerComposeLagoonLabel(composeVolumeValues.Labels, "lagoon.type")
		if volumeType == "" || volumeType == "none" {
			return nil, nil
		} else if volumeType == "persistent" {
			originalVolumeName := lagoon.GetVolumeNameFromComposeName(composeName, composeVolumeValues.Name)
			lagoonVolumeName := lagoon.GetLagoonVolumeName(originalVolumeName)
			volumeSize := lagoon.CheckDockerComposeLagoonLabel(composeVolumeValues.Labels, "lagoon.persistent.size")
			if volumeSize == "" {
				volumeSize = defaultAdditionalVolumeSize
			}
			// check the provided size is a valid resource size for kubernetes
			volS, err := ValidateResourceSize(volumeSize)
			if err != nil {
				return nil, fmt.Errorf("provided volume size for %s is not valid: %v", originalVolumeName, err)
			}
			// reject volumes over maxAdditionalVolumeSize for now
			maxSize, _ := ValidateResourceSize(maxAdditionalVolumeSize)
			if volS > maxSize {
				return nil, fmt.Errorf(
					"provided volume %s with size %s exceeds limit, if you need larger volumes please contact your Lagoon administrator",
					originalVolumeName,
					volumeSize,
				)
			}
			// create the volume values
			cVolume := &ComposeVolume{
				// use the lagoonVolumename which contains the `custom-` prefix
				Name: lagoonVolumeName,
				Size: volumeSize,
			}
			return cVolume, nil
		}
	}
	return nil, nil
}

// calculateServiceVolumes checks if a service type is allowed to have additional volumes attached, and if any volumes from docker compose
// are to be attached to this service or not
func calculateServiceVolumes(buildVolumes []ComposeVolume, lagoonType, servicePersistentName string, serviceLabels composetypes.Labels) ([]ServiceVolume, error) {
	serviceVolumes := []ServiceVolume{}
	if val, ok := servicetypes.ServiceTypes[lagoonType]; ok {
		for _, vol := range buildVolumes {
			volName := lagoon.GetVolumeNameFromLagoonVolume(vol.Name)
			// if there is a `lagoon.volumes.<xyx>.path` for a custom volume that matches the default persistent volume
			// don't add it again as a service volume
			if servicePersistentName != volName {
				volumePath := lagoon.CheckDockerComposeLagoonLabel(serviceLabels, fmt.Sprintf("lagoon.volumes.%s.path", volName))
				if volumePath != "" {
					if val.AllowAdditionalVolumes {
						sVol := ServiceVolume{
							ComposeVolume: vol,
							Path:          volumePath,
						}
						serviceVolumes = append(serviceVolumes, sVol)
					} else {
						// if the service type is not allowed additional volumes, return an error
						return nil, fmt.Errorf("the service type %s is not permitted to have additional volumes attached", lagoonType)
					}
				}
			}
		}
	}

	return serviceVolumes, nil
}
