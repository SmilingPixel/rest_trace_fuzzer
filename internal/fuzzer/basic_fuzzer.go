package fuzzer

import (
	"resttracefuzzer/internal/config"
	"resttracefuzzer/pkg/casemanager"
	"resttracefuzzer/pkg/static"
	"resttracefuzzer/pkg/utils"
	"time"

	"github.com/rs/zerolog/log"
)

// BasicFuzzer is a basic fuzzer, which is a simple implementation of the Fuzzer interface.
type BasicFuzzer struct {
	// APIManager is the API manager.
	APIManager *static.APIManager

	// CaseManager is the case manager.
	CaseManager *casemanager.CaseManager

	// Budget is the budget of the fuzzer, which is the maximum time the fuzzer can run, in milliseconds.
	Budget time.Duration

	// HTTPClient is the HTTP client.
	HTTPClient *utils.HTTPClient
}

// NewBasicFuzzer creates a new BasicFuzzer.
func NewBasicFuzzer(
	APIManager *static.APIManager,
	caseManager *casemanager.CaseManager,
) *BasicFuzzer {
	httpClient := &utils.HTTPClient{
		BaseURL: config.GlobalConfig.ServerBaseURL,
	}
	return &BasicFuzzer{
		APIManager:  APIManager,
		CaseManager: caseManager,
		Budget:      config.GlobalConfig.FuzzerBudget,
		HTTPClient:  httpClient,
	}
}

// Start starts the fuzzer.
func (f *BasicFuzzer) Start() error {
	// TODO: Implement this method. @xunzhou24

	startTime := time.Now()
	log.Info().Msgf("[BasicFuzzer.Start] Fuzzer started at %v", startTime)

	// loop:
	// 1. Pop a test case from the case manager.
	// 2. For each operation in the test case:
	//   a. Instantiate the operation.
	//   b. Make a request to the API.
	//   c. Process the response.
	// 3. Analyse the result, generate a report, and update the case manager.
	// 4. Go to step 1.
	for time.Since(startTime) <= f.Budget {
		testCase, err := f.CaseManager.Pop()
		if err != nil {
			log.Error().Msg("[BasicFuzzer.Start] Failed to pop a test case")
			break
		}

		err = f.ExecuteTestcase(testCase)
		if err != nil {
			log.Error().Msg("[BasicFuzzer.Start] Failed to execute the test case")
			break
		}

		// TODO: Analyse the result, generate a report, and update the case manager. @xunzhou24
	}

	return nil
}

func (f *BasicFuzzer) ExecuteTestcase(testcase *casemanager.Testcase) error {
	for _, operationCase := range testcase.OperationCases {
		operation := operationCase.Operation
		path := operationCase.APIMethod.Endpoint
		method := operationCase.APIMethod.Method
		statusCode, respBody, err := f.HTTPClient.PerformRequest(path, method)
		if err != nil {
			log.Error().Msg("[BasicFuzzer.ExecuteTestcase] Failed to perform request")
			return err
		}
		log.Info().Msgf("[BasicFuzzer.ExecuteTestcase] Response status code: %d, body: %s", statusCode, string(respBody))
	}
	return nil
}
