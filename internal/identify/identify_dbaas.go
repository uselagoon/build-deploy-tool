package identify

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
)

func IdentifyDBaaSConsumers(g generator.GeneratorInput) ([]string, error) {
	lagoonBuild, err := generator.NewGenerator(
		g,
	)
	if err != nil {
		return nil, err
	}
	ret := []string{}
	for _, svc := range lagoonBuild.BuildValues.Services {
		if svc.IsDBaaS || svc.IsSingle {
			ret = append(ret, fmt.Sprintf("%s:%s", svc.Name, svc.Type))
		}
	}
	return ret, nil
}
