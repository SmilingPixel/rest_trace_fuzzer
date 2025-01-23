package fuzzer

import (
	"resttracefuzzer/internal/config"
	"resttracefuzzer/pkg/casemanager"
	"resttracefuzzer/pkg/feedback"
	"resttracefuzzer/pkg/static"
	"resttracefuzzer/pkg/utils"
	"time"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
)

// BasicFuzzer is a basic fuzzer, which is a simple implementation of the Fuzzer interface.
type BasicFuzzer struct {
	// APIManager is the API manager.
	APIManager *static.APIManager

	// CaseManager is the case manager.
	CaseManager *casemanager.CaseManager

	// ResponseChecker checks the response, and records the hit count of the status code.
	ResponseChecker *feedback.ResponseChecker

	// TraceManager manages traces.
	TraceManager *feedback.TraceManager

	// RuntimeGraph is the runtime graph, including coverage information.
	RunTimeGraph *feedback.RuntimeGraph

	// Budget is the budget of the fuzzer, which is the maximum time the fuzzer can run, in milliseconds.
	Budget time.Duration

	// HTTPClient is the HTTP client.
	HTTPClient *utils.HTTPClient
}

// NewBasicFuzzer creates a new BasicFuzzer.
func NewBasicFuzzer(
	APIManager *static.APIManager,
	caseManager *casemanager.CaseManager,
	responseChecker *feedback.ResponseChecker,
	traceManager *feedback.TraceManager,
	runtimeGraph *feedback.RuntimeGraph,
) *BasicFuzzer {
	httpClient := utils.NewHTTPClient(
		config.GlobalConfig.ServerBaseURL,
	)
	return &BasicFuzzer{
		APIManager:      APIManager,
		CaseManager:     caseManager,
		ResponseChecker: responseChecker,
		TraceManager:    traceManager,
		Budget:          config.GlobalConfig.FuzzerBudget,
		HTTPClient:      httpClient,
		RunTimeGraph:    runtimeGraph,
	}
}

// Start starts the fuzzer.
func (f *BasicFuzzer) Start() error {
	// TODO: Implement this method. @xunzhou24

	startTime := time.Now()
	log.Info().Msgf("[BasicFuzzer.Start] Fuzzer started at %v", startTime)

	// loop:
	// 1. Pop a test scenario from the case manager.
	// 2. For each operation in the test scenario:
	//   a. Instantiate the operation.
	//   b. Make a request to the API.
	//   c. Process the response.
	// 3. Analyse the result, generate a report, and update the case manager.
	// 4. Go to step 1.
	for time.Since(startTime) <= f.Budget {
		testScenario, err := f.CaseManager.Pop()
		if err != nil {
			log.Error().Err(err).Msg("[BasicFuzzer.Start] Failed to pop a test scenario")
			break
		}

		err = f.ExecuteTestScenario(testScenario)
		if err != nil {
			log.Error().Err(err).Msg("[BasicFuzzer.Start] Failed to execute the test scenario")
			break
		}

		// TODO: Analyse the result, generate a report, and update the case manager. @xunzhou24
	}

	log.Info().Msg("[BasicFuzzer.Start] Fuzzer stopped")
	return nil
}

func (f *BasicFuzzer) ExecuteTestScenario(testScenario *casemanager.TestScenario) error {
	for _, operationCase := range testScenario.OperationCases {
		// TODO: pass body and params @xunzhou24
		err := f.ExecuteCaseOperation(operationCase)
		if err != nil {
			log.Error().Err(err).Msg("[BasicFuzzer.ExecuteTestcase] Failed to execute operation")
		}
		statusCode := operationCase.ResponseStatusCode

		// Check the response.
		err = f.ResponseChecker.CheckResponse(operationCase.APIMethod, statusCode)
		if err != nil {
			log.Error().Err(err).Msg("[BasicFuzzer.ExecuteTestcase] Failed to check response")
			return err
		}

		// TODO: parse and validate the response body @xunzhou24

		// fetch traces from the service, parse them, and update local runtime graph.
		err = f.TraceManager.PullTraces()
		if err != nil {
			log.Error().Err(err).Msg("[BasicFuzzer.ExecuteTestcase] Failed to pull traces")
			return err
		}
		callInfoList, err := f.TraceManager.GetCallInfos(nil) // TODO: pass the trace @xunzhou24
		if err != nil {
			log.Error().Err(err).Msg("[BasicFuzzer.ExecuteTestcase] Failed to get call infos")
			return err
		}
		err = f.RunTimeGraph.UpdateFromCallInfos(callInfoList)
		if err != nil {
			log.Error().Err(err).Msg("[BasicFuzzer.ExecuteTestcase] Failed to update runtime graph")
			return err
		}
		log.Info().Msg("[BasicFuzzer.ExecuteTestcase] Operation executed successfully")
	}
	return nil
}

// ExecuteCaseOperation executes a case operation from a test case.
// This method makes HTTP call, and fills the response in the operation case.
func (f *BasicFuzzer) ExecuteCaseOperation(operationCase *casemanager.OperationCase) error {
	// operation := operationCase.Operation
	path := operationCase.APIMethod.Endpoint
	method := operationCase.APIMethod.Method
	log.Debug().Msgf("[BasicFuzzer.ExecuteCaseOperation] Execute operation: %s %s", method, path)
	statusCode, respBodyBytes, err := f.HTTPClient.PerformRequest(path, method, nil, nil, nil)
	if err != nil {
		// A failed request will not stop the fuzzing process.
		log.Error().Err(err).Msg("[BasicFuzzer.ExecuteCaseOperation] Failed to perform request")
	}
	respBody := make(map[string]interface{})
	err = sonic.Unmarshal(respBodyBytes, &respBody)
	if err != nil {
		// A broken response body will not stop the fuzzing process.
		log.Error().Err(err).Msg("[BasicFuzzer.ExecuteCaseOperation] Failed to unmarshal response body")
	}
	// Fill the response in the operation case.
	operationCase.ResponseStatusCode = statusCode
	operationCase.ResponseBody = respBody
	log.Debug().Msgf("[BasicFuzzer.ExecuteTestcase] Response status code: %d, body: %s", statusCode, string(respBodyBytes))
	return nil
}

// GetRuntimeGraph gets the runtime graph.
func (f *BasicFuzzer) GetRuntimeGraph() *feedback.RuntimeGraph {
	return f.RunTimeGraph
}
