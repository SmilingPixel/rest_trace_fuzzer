package fuzzer

import (
	"resttracefuzzer/pkg/apimanager"

	"github.com/cloudwego/hertz/pkg/app/client"
)

type Fuzzer struct {
	Client *client.Client
	APIManager *apimanager.APIManager 
}

// NewFuzzerClient creates a new FuzzerClient.
func NewFuzzer(APIManager *apimanager.APIManager) *Fuzzer {
	c, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	return &Fuzzer{
		Client: c,
		APIManager: APIManager,
	}
}

