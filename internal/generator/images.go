package generator

import (
	"errors"
	"fmt"
	"os"
	"strings"

	composetypes "github.com/compose-spec/compose-go/types"
	"github.com/drone/envsubst"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

// generateImageBuild will work out all the build arguments and build options required for a specific service from docker compose
func generateImageBuild(buildValues BuildValues, composeServiceValues composetypes.ServiceConfig, composeService string) (ImageBuild, error) {
	// create a holder for all the docker related information, if this is a pull through image or a build image
	imageBuild := ImageBuild{}
	// if this is not a promote environment, then attempt to work out the image build information that is required for the builder
	if buildValues.BuildType != "promote" {
		// handle extracting the built image name from the provided image references
		if composeServiceValues.Build != nil {
			// if a build spec is defined, consume it
			// set the dockerfile
			imageBuild.DockerFile = composeServiceValues.Build.Dockerfile
			// set the context if found, otherwise set '.'
			imageBuild.Context = func(s string) string {
				if s == "" {
					return "."
				}
				return s
			}(composeServiceValues.Build.Context)
			// if there is a build target defined, set that here too
			imageBuild.Target = composeServiceValues.Build.Target
		}
		// if there is a dockerfile defined in the
		if buildValues.LagoonYAML.Environments[buildValues.Environment].Overrides[composeService].Build.Dockerfile != "" {
			imageBuild.DockerFile = buildValues.LagoonYAML.Environments[buildValues.Environment].Overrides[composeService].Build.Dockerfile
			if imageBuild.Context == "" {
				// if we get here, it means that a dockerfile override was defined in the .lagoon.yml file
				// but there was no `build` spec defined in the docker-compose file, so this just sets the context to the default `.`
				// in the same way the legacy script used to do it
				imageBuild.Context = "."
			}
		}
		if imageBuild.DockerFile == "" {
			// no dockerfile determined, this must be a pull through image
			if composeServiceValues.Image == "" {
				return imageBuild, fmt.Errorf(
					"no defined Dockerfile or Image for service %s", composeService,
				)
			}
			// check docker-compose override image
			pullImage := lagoon.CheckDockerComposeLagoonLabel(composeServiceValues.Labels, "lagoon.image")
			// check lagoon.yml override image
			if buildValues.LagoonYAML.Environments[buildValues.Environment].Overrides[composeService].Image != "" {
				pullImage = buildValues.LagoonYAML.Environments[buildValues.Environment].Overrides[composeService].Image
			}
			if pullImage != "" {
				// if an override image is provided, envsubst it
				// not really sure why we do this, but legacy bash says `expand environment variables from ${OVERRIDE_IMAGE}`
				// so there may be some undocumented functionality that allows people to use envvars in their image overrides?
				evalImage, err := envsubst.EvalEnv(pullImage)
				if err != nil {
					return imageBuild, fmt.Errorf(
						"error evaluating override image %s with envsubst", pullImage,
					)
				}
				// set the evalled image now
				pullImage = evalImage
			} else {
				// else set the pullimage to whatever is defined in the docker-compose file otherwise
				pullImage = composeServiceValues.Image
			}
			// if the image just is an image name (like "alpine") we prefix it with `libary/` as the imagecache does not understand
			// the magic `alpine` image
			if !strings.Contains(pullImage, "/") {
				imageBuild.PullImage = fmt.Sprintf("library/%s", pullImage)
			} else {
				imageBuild.PullImage = pullImage
			}
			if !ContainsRegistry(buildValues.ContainerRegistry, pullImage) {
				// if the image isn't in dockerhub, then the imagecache can't be used
				if buildValues.ImageCache != "" && strings.Count(pullImage, "/") == 1 && !buildValues.IgnoreImageCache {
					imageBuild.PullImage = fmt.Sprintf("%s%s", buildValues.ImageCache, imageBuild.PullImage)
				}
			}
		} else {
			// otherwise this must be an image build
			// set temporary image to prevent clashes?? not sure this is even required, the temporary name is just as unique as the final image name eventually is
			// so clashing would occur in both situations
			imageBuild.TemporaryImage = fmt.Sprintf("%s-%s", buildValues.Namespace, composeService) //@TODO maybe get rid of this
			if buildValues.LagoonYAML.Environments[buildValues.Environment].Overrides[composeService].Build.Context != "" {
				imageBuild.Context = buildValues.LagoonYAML.Environments[buildValues.Environment].Overrides[composeService].Build.Context
			}
			// check the dockerfile exists
			if _, err := os.Stat(fmt.Sprintf("%s/%s", imageBuild.Context, imageBuild.DockerFile)); errors.Is(err, os.ErrNotExist) {
				return imageBuild, fmt.Errorf(
					"defined Dockerfile %s for service %s not found",
					fmt.Sprintf("%s/%s", imageBuild.Context, imageBuild.DockerFile), composeService,
				)
			}
		}
	}
	// since we know what the final build image will be, we can set it here, this is what all images will be built as during the build
	// for `pullimages` they will get retagged as this imagename and pushed to the registry
	imageBuild.BuildImage = fmt.Sprintf("%s/%s/%s/%s:%s", buildValues.ImageRegistry, buildValues.Project, buildValues.Environment, composeService, "latest")
	if buildValues.BuildType == "promote" {
		imageBuild.PromoteImage = fmt.Sprintf("%s/%s/%s/%s:%s", buildValues.ImageRegistry, buildValues.Project, buildValues.PromotionSourceEnvironment, composeService, "latest")
	}
	// populate the docker derived information here, this information will be used by the build and pushing scripts
	return imageBuild, nil
	// cService.ImageBuild = &imageBuild

	// // // cService.ImageName = buildValues.ImageReferences[composeService]
	// unfortunately, this uses a specific hash which is computed "AFTER" the image builds take place, so this
	// `ImageName` is an unreliable field in respect to consuming data from generator during phases of a build
	// it would be great if there was a way to precalculate this, but there are other issues that could pop up
	// using the buildname as the tag could be one way, but this could result in container restarts even if the image hash does not change :(
	// for now `ImageName` is disabled, and ImageReferences must be provided whenever templating occurs that needs an image reference
	// luckily the templating engine will reproduce identical data when it is run as to when the image build data is populated as above
	// so when the templating is done in a later step, at least it can be informed of the resulting image references by way of the
	// images flag that is passed to it
	/*
		// example in code
		ImageReferences: map[string]string{
			"myservice": "harbor.example.com/example-project/environment-name/myservice@sha256:abcdefg",
		},

		// in bash, the images are provided from yaml as base64 encoded data to retain formatting
		// the command `lagoon-services` decodes and unmarshals it
		build-deploy-tool template lagoon-services \
			--saved-templates-path ${LAGOON_SERVICES_YAML_FOLDER} \
			--images $(yq3 r -j /kubectl-build-deploy/images.yaml | jq -M -c | base64 -w0)
	*/
}
