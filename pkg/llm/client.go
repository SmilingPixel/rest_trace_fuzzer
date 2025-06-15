package llm

import "context"


// LLMClient defines an interface for interacting with a Large Language Model (LLM) service.
// It provides a method to generate chat-based completions based on a given prompt.
//
// This interface can be implemented by any client that communicates with an LLM API,
// enabling flexibility and abstraction for different LLM providers.
//
// Methods:
//   - ChatCompletion: Generates a response from the LLM based on the provided prompt.
//
// Example usage:
//   var client LLMClient
//   response, err := client.ChatCompletion(ctx, "Hello, how are you?")
//   if err != nil {
//       log.Fatalf("Error generating chat completion: %v", err)
//   }
//   fmt.Println("LLM Response:", response)
type LLMClient interface {
	ChatCompletion(ctx context.Context, prompt string) (string, error)
}