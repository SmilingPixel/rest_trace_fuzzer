package llm

import (
	"context"

	"github.com/openai/openai-go" // imported as openai
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
	"github.com/rs/zerolog/log"
)

// OpenAIClient is a client for interacting with the OpenAI API.
type OpenAIClient struct {
	client *openai.Client   // The underlying OpenAI client.
	model  shared.ChatModel // The model to be used for chat completions.
}

// NewOpenAIClient creates a new instance of OpenAIClient.
// Parameters:
// - baseURL: The base URL of the OpenAI API.
// - APIKey: The API key for authenticating with the OpenAI API.
// - model: The chat model to be used for generating completions.
// Returns an instance of OpenAIClient.
func NewOpenAIClient(
	baseURL string,
	APIKey string,
	model shared.ChatModel,
) *OpenAIClient {
	client := openai.NewClient(
		option.WithAPIKey(APIKey),
		option.WithBaseURL(baseURL),
	)
	return &OpenAIClient{
		client: &client,
	}
}

// ChatCompletion generates a chat completion for the given prompt using the OpenAI API.
// Parameters:
// - ctx: The context for the API request.
// - prompt: The input prompt for generating the chat completion.
// Returns the generated chat completion as a string, or an error if the request fails.
func (c *OpenAIClient) ChatCompletion(ctx context.Context, prompt string) (string, error) {
	chatCompletion, err := c.client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model: c.model,
	})
	if err != nil {
		log.Err(err).Msg("[OpenAIClient.ChatCompletion] Error generating chat completion")
		return "", err
	}
	log.Info().Msgf("[OpenAIClient.ChatCompletion] Generated chat completion: %s", chatCompletion.Choices[0].Message.Content)
	return chatCompletion.Choices[0].Message.Content, nil
}
