package main

import (
	"context"

	"github.com/maksymenkoml/lingoose/legacy/prompt"
	llmmock "github.com/maksymenkoml/lingoose/llm/mock"
)

func main() {

	prompt := prompt.New("How are you?")

	llm := llmmock.LlmMock{}

	output, err := llm.Completion(context.Background(), prompt.String())
	if err != nil {
		panic(err)
	}

	println(output)
}
