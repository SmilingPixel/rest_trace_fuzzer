package report

import (
	"fmt"
	"os"
	"resttracefuzzer/pkg/resource"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
)

type FuzzerStateReporter struct {
}

func NewFuzzerStateReporter() *FuzzerStateReporter {
	return &FuzzerStateReporter{}
}

func (r *FuzzerStateReporter) GenerateFuzzerStateReport(resourceManager *resource.ResourceManager, outputPath string) error {
	if resourceManager == nil {
		log.Error().Msg("[FuzzerStateReporter.GenerateFuzzerStateReport] resourceManager is nil.")
		return fmt.Errorf("resourceManager is nil")
		
	}

	// We do not jsonify the resources directly, as we want to use custom ToJSONObject() method of Resource.
	resourceNameMap := resourceManager.ResourceNameMap
	if resourceNameMap == nil {
		log.Error().Msg("[FuzzerStateReporter.GenerateFuzzerStateReport] resourceNameMap is nil.")
		return fmt.Errorf("resourceNameMap is nil")
	}
	resourceJSONObjectNameMap := make(map[string][]interface{})
	for resrcName, resrcList := range resourceNameMap {
		resourceJSONObjectNameMap[resrcName] = make([]interface{}, 0)
		for _, resrc := range resrcList {
			resourceJSONObjectNameMap[resrcName] = append(resourceJSONObjectNameMap[resrcName], resrc.ToJSONObject())
		}
	}

	fuzzerStateReport := FuzzerStateReport{
		ResourceNameMap:            resourceManager.ResourceNameMap,
		ResourceJSONObjectNameMap:   resourceJSONObjectNameMap,
	}
	reportBytes, err := sonic.Marshal(fuzzerStateReport)
	if err != nil {
		log.Err(err).Msg("[FuzzerStateReporter.GenerateFuzzerStateReport] Failed to marshal the report")
		return err
	}

	err = os.WriteFile(outputPath, reportBytes, 0644)
	if err != nil {
		log.Err(err).Msgf("[FuzzerStateReporter.GenerateFuzzerStateReport] Failed to write the fuzzer state report to file")
		return err
	}
	log.Info().Msgf("[FuzzerStateReporter.GenerateFuzzerStateReport] Fuzzer state report has been written to %s", outputPath)
	return nil
}
