package fuzzer

import (
	"resttracefuzzer/internal/config"
	"resttracefuzzer/pkg/casemanager"
	"resttracefuzzer/pkg/feedback"
	"resttracefuzzer/pkg/feedback/trace"
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
	TraceManager *trace.TraceManager

	// RuntimeGraph is the runtime graph, including coverage information.
	RunTimeGraph *feedback.RuntimeGraph

	// Budget is the budget of the fuzzer, which is the maximum time the fuzzer can run, in milliseconds.
	Budget time.Duration

	// HTTPClient is the HTTP client.
	HTTPClient *utils.HTTPClient

	// FuzzingSnapshot is the snapshot of the fuzzing process.
	FuzzingSnapshot *FuzzingSnapshot
}

// NewBasicFuzzer creates a new BasicFuzzer.
func NewBasicFuzzer(
	APIManager *static.APIManager,
	caseManager *casemanager.CaseManager,
	responseChecker *feedback.ResponseChecker,
	traceManager *trace.TraceManager,
	runtimeGraph *feedback.RuntimeGraph,
) *BasicFuzzer {
	httpClient := utils.NewHTTPClient(
		config.GlobalConfig.ServerBaseURL,
		[]string{config.GlobalConfig.TraceIDHeaderKey},
	)
	fuzzingSnapshot := NewFuzzingSnapshot()
	return &BasicFuzzer{
		APIManager:      APIManager,
		CaseManager:     caseManager,
		ResponseChecker: responseChecker,
		TraceManager:    traceManager,
		Budget:          time.Duration(config.GlobalConfig.FuzzerBudget) * time.Second, // Convert seconds to nanoseconds.
		HTTPClient:      httpClient,
		RunTimeGraph:    runtimeGraph,
		FuzzingSnapshot: fuzzingSnapshot,
	}
}

// Start starts the fuzzer.
// The fuzzer will run until the budget is exhausted or some error occurs.
func (f *BasicFuzzer) Start() error {

	startTime := time.Now()
	log.Info().Msgf("[BasicFuzzer.Start] Fuzzer started at %v, Budget: %v", startTime, f.Budget)

	// loop:
	// 1. Pop a test scenario from the case manager.
	// 2. For each operation in the test scenario:
	//   a. Instantiate the operation.
	//   b. Make a request to the API.
	//   c. Process the response.
	// 3. Analyse the result, generate a report, and update the case manager.
	// 4. Go to step 1.
	for time.Since(startTime) <= f.Budget {
		testScenario, err := f.CaseManager.PopAndPopulate()
		if err != nil {
			log.Err(err).Msg("[BasicFuzzer.Start] Failed to pop a test scenario")
			break
		}

		err = f.ExecuteTestScenario(testScenario)
		if err != nil {
			log.Err(err).Msg("[BasicFuzzer.Start] Failed to execute the test scenario")
			break
		}
	}

	log.Info().Msg("[BasicFuzzer.Start] Fuzzer stopped")
	return nil
}

// ExecuteTestScenario executes a test scenario (a sequence of operation cases).
// This method makes HTTP calls, checks the response, and updates the runtime graph.
// If the analysers conclude that the test scenario is interesting, the case manager will be updated (e.g., add the test scenario back to queue).
func (f *BasicFuzzer) ExecuteTestScenario(testScenario *casemanager.TestScenario) error {
	for _, operationCase := range testScenario.OperationCases {
		err := f.ExecuteCaseOperation(operationCase)
		if err != nil {
			log.Err(err).Msg("[BasicFuzzer.ExecuteTestScenario] Failed to execute operation")
		}
		statusCode := operationCase.ResponseStatusCode

		// Check the response.
		err = f.ResponseChecker.CheckResponse(operationCase.APIMethod, statusCode)
		if err != nil {
			log.Err(err).Msg("[BasicFuzzer.ExecuteTestScenario] Failed to check response")
			return err
		}

		// TODO: parse and validate the response body @xunzhou24

		// fetch the trace from the service, parse it, and update local runtime graph.
		traceID, exist := operationCase.ResponseHeaders[config.GlobalConfig.TraceIDHeaderKey]
		if !exist {
			log.Warn().Msg("[BasicFuzzer.ExecuteTestScenario] No trace ID found in the response headers")
			continue
		}
		newTrace, err := f.TraceManager.PullTraceByIDAndReturn(traceID)
		if err != nil {
			log.Err(err).Msg("[BasicFuzzer.ExecuteTestScenario] Failed to pull traces")
			return err
		}
		// During the conversion, spans of kind 'internal' would be ignored, as we only care about the calls between services.
		callInfoList, err := f.TraceManager.ConvertTraces2CallInfos([]*trace.SimplifiedTrace{newTrace})
		if err != nil {
			log.Err(err).Msg("[BasicFuzzer.ExecuteTestScenario] Failed to get call infos")
			return err
		}
		err = f.RunTimeGraph.UpdateFromCallInfos(callInfoList)
		if err != nil {
			log.Err(err).Msg("[BasicFuzzer.ExecuteTestScenario] Failed to update runtime graph")
			return err
		}
		log.Info().Msg("[BasicFuzzer.ExecuteTestScenario] Operation executed successfully")
	}

	hasAchieveNewCoverage := f.FuzzingSnapshot.Update(
		f.RunTimeGraph.GetEdgeCoverage(),
		f.ResponseChecker.GetCoveredStatusCodeCount(),
	)
	log.Info().Msgf("[BasicFuzzer.ExecuteTestScenario] Finish execute current test scenario, Edge coverage: %f, covered status code count: %d, hasAchieveNewCoverage: %v", f.RunTimeGraph.GetEdgeCoverage(), f.ResponseChecker.GetCoveredStatusCodeCount(), hasAchieveNewCoverage)

	// Pass the scenario and the result back to the case manager,
	// and decide whether to put the scenario back to the queue and to generate a new one.
	err := f.CaseManager.EvaluateScenarioAndTryUpdate(hasAchieveNewCoverage, testScenario)
	if err != nil {
		log.Err(err).Msg("[BasicFuzzer.ExecuteTestScenario] Failed to evaluate scenario and try update")
		return err
	}

	return nil
}

// ExecuteCaseOperation executes a case operation from a test case.
// This method makes HTTP call, and fills the response in the operation case.
func (f *BasicFuzzer) ExecuteCaseOperation(operationCase *casemanager.OperationCase) error {
	path := operationCase.APIMethod.Endpoint
	method := operationCase.APIMethod.Method
	headers := operationCase.RequestHeaders
	pathParams := operationCase.RequestPathParams
	queryParams := operationCase.RequestQueryParams
	body := operationCase.RequestBody
	log.Debug().Msgf("[BasicFuzzer.ExecuteCaseOperation] Execute operation: %s %s", method, path)
	statusCode, headers, respBodyBytes, err := f.HTTPClient.PerformRequest(path, method, headers, pathParams, queryParams, body)
	if err != nil {
		// A failed request will not stop the fuzzing process.
		log.Err(err).Msg("[BasicFuzzer.ExecuteCaseOperation] Failed to perform request")
	}
	// Unmarshal the response body only if the status code is successful, otherwise keep the response body as bytes.
	var respBody interface{}
	if utils.IsStatusCodeSuccess(statusCode) {
		// Unmarshal the response body.
		err = sonic.Unmarshal(respBodyBytes, &respBody)
		if err != nil {
			// A broken response body will not stop the fuzzing process.
			log.Err(err).Msg("[BasicFuzzer.ExecuteCaseOperation] Failed to unmarshal response body")
		}
	} else {
		log.Info().Msgf("[BasicFuzzer.ExecuteCaseOperation] Unsuccessful response status code: %d, response body: %s", statusCode, string(respBodyBytes))
		respBody = respBodyBytes
	}
	// Fill the response in the operation case.
	operationCase.ResponseStatusCode = statusCode
	operationCase.ResponseHeaders = headers
	operationCase.ResponseBody = respBody
	log.Debug().Msgf("[BasicFuzzer.ExecuteCaseOperation] Response status code: %d, body: %s", statusCode, string(respBodyBytes))
	return nil
}

// GetRuntimeGraph gets the runtime graph.
func (f *BasicFuzzer) GetRuntimeGraph() *feedback.RuntimeGraph {
	return f.RunTimeGraph
}
