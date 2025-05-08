package main

import (
	"context"
	"fmt"

	openaiembedder "github.com/maksymenkoml/lingoose/embedder/openai"
	"github.com/maksymenkoml/lingoose/index"
	"github.com/maksymenkoml/lingoose/index/vectordb/jsondb"
	"github.com/maksymenkoml/lingoose/llm/cache"
	"github.com/maksymenkoml/lingoose/llm/openai"
	"github.com/maksymenkoml/lingoose/thread"
)

func main() {

	llm := openai.New().WithCache(
		cache.New(
			index.New(
				jsondb.New().WithPersist("index.json"),
				openaiembedder.New(openaiembedder.AdaEmbeddingV2),
			),
		).WithTopK(3),
	)

	questions := []string{
		"what's github",
		"can you explain what GitHub is",
		"can you tell me more about GitHub",
		"what is the purpose of GitHub",
	}

	for _, question := range questions {
		t := thread.New().AddMessage(
			thread.NewUserMessage().AddContent(
				thread.NewTextContent(question),
			),
		)

		err := llm.Generate(context.Background(), t)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println(t)
	}
}
