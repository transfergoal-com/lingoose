package main

import (
	"context"
	"fmt"
	"os"

	openaiembedder "github.com/maksymenkoml/lingoose/embedder/openai"
	"github.com/maksymenkoml/lingoose/index"
	"github.com/maksymenkoml/lingoose/index/vectordb/jsondb"
	"github.com/maksymenkoml/lingoose/linglet/qa"
	"github.com/maksymenkoml/lingoose/llm/openai"
)

// download https://raw.githubusercontent.com/hwchase17/chat-your-data/master/state_of_the_union.txt

func main() {
	qa := qa.New(
		openai.New().WithTemperature(0),
		index.New(
			jsondb.New().WithPersist("db.json"),
			openaiembedder.New(openaiembedder.AdaEmbeddingV2),
		),
	)

	_, err := os.Stat("db.json")
	if os.IsNotExist(err) {
		err = qa.AddSource(context.Background(), "state_of_the_union.txt")
		if err != nil {
			panic(err)
		}
	}

	response, err := qa.Run(context.Background(), "What is the NATO purpose?")
	if err != nil {
		panic(err)
	}

	fmt.Println(response)
}
