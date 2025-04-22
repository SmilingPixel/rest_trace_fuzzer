package fuzzer

import (
	"resttracefuzzer/internal/config"
	"resttracefuzzer/pkg/casemanager"
	"resttracefuzzer/pkg/feedback"
	"resttracefuzzer/pkg/feedback/trace"
	"resttracefuzzer/pkg/report"
	fuzzruntime "resttracefuzzer/pkg/runtime"
	"resttracefuzzer/pkg/static"
	"resttracefuzzer/pkg/utils/http"
	"time"

	hertzclient "github.com/cloudwego/hertz/pkg/app/client"
	"github.com/rs/zerolog/log"
)

// BasicFuzzer is a basic fuzzer, which is a simple implementation of the Fuzzer interface.
type BasicFuzzer struct {
	// APIManager is the API manager.
	APIManager *static.APIManager

	// CaseManager is the case manager.
	CaseManager *casemanager.CaseManager

	// ResponseProcesser checks and processes the response.
	ResponseProcesser *feedback.ResponseProcesser

	// TraceManager manages traces.
	TraceManager *trace.TraceManager

	// CallInfoGraph is the runtime graph of call info, including coverage information.
	CallInfoGraph *fuzzruntime.CallInfoGraph

	// ReachabilityMap stores the reachability (from external APIs to internal service interfaces) information of the API.
	ReachabilityMap *fuzzruntime.RuntimeReachabilityMap

	// Budget is the budget of the fuzzer, which is the maximum time the fuzzer can run, in milliseconds.
	Budget time.Duration

	// HTTPClient is the HTTP client.
	HTTPClient *http.HTTPClient

	// FuzzingSnapshot is the snapshot of the fuzzing process.
	FuzzingSnapshot *FuzzingSnapshot

	// TestLogReporter is responsible for logging the tested operations (with their results),
	// and generating a report after the fuzzing process.
	TestLogReporter *report.TestLogReporter
}

// NewBasicFuzzer creates a new BasicFuzzer.
func NewBasicFuzzer(
	APIManager *static.APIManager,
	caseManager *casemanager.CaseManager,
	responseProcesser *feedback.ResponseProcesser,
	traceManager *trace.TraceManager,
	callInfoGraph *fuzzruntime.CallInfoGraph,
	reachabilityMap *fuzzruntime.RuntimeReachabilityMap,
	testLogReporter *report.TestLogReporter,
) *BasicFuzzer {
	httpClientMiddles := make([]http.HTTPClientMiddleware, 0)
	if config.GlobalConfig.HTTPMiddlewareScriptPath != "" {
		middleware := http.NewHTTPClientScriptMiddleware(config.GlobalConfig.HTTPMiddlewareScriptPath)
		if middleware != nil {
			httpClientMiddles = append(httpClientMiddles, middleware)
		}
	}
	httpClient := http.NewHTTPClient(
		config.GlobalConfig.ServerBaseURL,
		[]string{config.GlobalConfig.TraceIDHeaderKey},
		httpClientMiddles,
		hertzclient.WithDialTimeout(time.Duration(config.GlobalConfig.HTTPClientDialTimeout) * time.Second),
	)
	fuzzingSnapshot := NewFuzzingSnapshot()

	// If budget is not positive, no fuzzing will be performed.
	// Just give a warning and continue, as the user may want a dry run or a initial data flow graph.
	if config.GlobalConfig.FuzzerBudget <= 0 {
		log.Warn().Msg("[BasicFuzzer.NewBasicFuzzer] Fuzzer budget is not positive, no fuzzing will be performed")
	}
	
	return &BasicFuzzer{
		APIManager:        APIManager,
		CaseManager:       caseManager,
		ResponseProcesser: responseProcesser,
		TraceManager:      traceManager,
		Budget:            time.Duration(config.GlobalConfig.FuzzerBudget) * time.Second, // Convert seconds to nanoseconds.
		HTTPClient:        httpClient,
		CallInfoGraph:     callInfoGraph,
		ReachabilityMap:   reachabilityMap,
		FuzzingSnapshot:   fuzzingSnapshot,
		TestLogReporter:   testLogReporter,
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
// This method makes HTTP calls, processes the response, and updates the runtime call info graph.
// If the analysers conclude that the test scenario or its test operation cases are interesting, the case manager will be updated (e.g., mutate the test scenario and add it back to queue).
func (f *BasicFuzzer) ExecuteTestScenario(testScenario *casemanager.TestScenario) error {
	var hasScenarioAchieveNewCoverage bool
	for _, operationCase := range testScenario.OperationCases {
		// If error occurs during execution of the operation case, stop the whole test scenario.
		// Otherwise, continue to the next operation case.
		err := f.ExecuteCaseOperation(operationCase)
		if err != nil {
			log.Err(err).Msg("[BasicFuzzer.ExecuteTestScenario] Failed to execute operation")
			return err
		}
		statusCode := operationCase.ResponseStatusCode
		responseBody := operationCase.ResponseBody

		// Process the response.
		// This phase would check the response status code and response body.
		// The body would be stored in the resource manager if the request is successful.
		// Error in processing the response will not stop the fuzzing process.
		err = f.ResponseProcesser.ProcessResponse(operationCase.APIMethod, statusCode, responseBody)
		if err != nil {
			log.Err(err).Msg("[BasicFuzzer.ExecuteTestScenario] Failed to process response")
			continue // continue to the next operation case instead of stopping the fuzzing process
		}

		// fetch the trace from the service, parse it, and update local runtime call info graph.
		traceID, exist := operationCase.ResponseHeaders[config.GlobalConfig.TraceIDHeaderKey]
		if !exist || traceID == "" {
			log.Warn().Msg("[BasicFuzzer.ExecuteTestScenario] No trace ID found in the response headers")
			continue
		}
		newTrace, err := f.TraceManager.PullTraceByIDAndReturn(traceID)
		if err != nil {
			log.Err(err).Msg("[BasicFuzzer.ExecuteTestScenario] Failed to pull traces")
			continue
		}
		// During the conversion, spans of kind 'internal' would be ignored, as we only care about the calls between services.
		callInfoList, err := f.TraceManager.BatchConvertTrace2CallInfos([]*trace.SimplifiedTrace{newTrace})
		if err != nil {
			log.Err(err).Msg("[BasicFuzzer.ExecuteTestScenario] Failed to get call infos")
			continue
		}

		// Update runtime info, including call info graph and reachability map.
		err = f.CallInfoGraph.UpdateFromCallInfos(callInfoList)
		if err != nil {
			log.Err(err).Msg("[BasicFuzzer.ExecuteTestScenario] Failed to update runtime call info graph")
			continue
		}
		err = f.ReachabilityMap.UpdateFromCallInfos(operationCase.APIMethod, callInfoList)
		if err != nil {
			log.Err(err).Msg("[BasicFuzzer.ExecuteTestScenario] Failed to update reachability map")
			continue
		}

		log.Info().Msg("[BasicFuzzer.ExecuteTestScenario] Operation executed successfully")

		hasOperationAchieveNewCoverage := f.FuzzingSnapshot.Update(
			f.CallInfoGraph.GetEdgeCoverage(),
			f.ResponseProcesser.GetCoveredStatusCodeCount(),
		)
		hasScenarioAchieveNewCoverage = hasScenarioAchieveNewCoverage || hasOperationAchieveNewCoverage

		// Pass the operation and the its execution result back to the case manager,
		// and:
		//  1. decide whether its operation cases are interesting or not (i.e., update their energy)
		//  2. may mutate the operation cases and add them to the operation case queue.
		err = f.CaseManager.EvaluateOperationCaseAndTryUpdate(hasOperationAchieveNewCoverage, operationCase)
		if err != nil {
			log.Err(err).Msg("[BasicFuzzer.ExecuteTestScenario] Failed to evaluate operation and try update")
			return err
		}
	}

	log.Info().Msgf("[BasicFuzzer.ExecuteTestScenario] Finish execute current test scenario (UUID: %s), Edge coverage: %f, covered status code count: %d, hasScenarioAchieveNewCoverage: %v", testScenario.UUID.String(), f.CallInfoGraph.GetEdgeCoverage(), f.ResponseProcesser.GetCoveredStatusCodeCount(), hasScenarioAchieveNewCoverage)

	// Pass the scenario and the result back to the case manager,
	// and:
	//  1. decide whether the scenario is interesting or not (i.e., update its energy)
	//  2. may mutate the scenario and add it back to the scenario queue.
	err := f.CaseManager.EvaluateScenarioAndTryUpdate(hasScenarioAchieveNewCoverage, testScenario)
	if err != nil {
		log.Err(err).Msg("[BasicFuzzer.ExecuteTestScenario] Failed to evaluate scenario and try update")
		return err
	}

	// Log the tested scenario.
	f.TestLogReporter.LogTestScenario(testScenario)

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

	// Fill the response in the operation case.
	operationCase.ResponseStatusCode = statusCode
	operationCase.ResponseHeaders = headers
	operationCase.ResponseBody = respBodyBytes
	log.Debug().Msgf("[BasicFuzzer.ExecuteCaseOperation] Response status code: %d, body: %s", statusCode, string(respBodyBytes))
	return nil
}

// GetCallInfoGraph gets the runtime call info graph.
func (f *BasicFuzzer) GetCallInfoGraph() *fuzzruntime.CallInfoGraph {
	return f.CallInfoGraph
}
