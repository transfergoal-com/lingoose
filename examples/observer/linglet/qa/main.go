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
	"github.com/maksymenkoml/lingoose/observer"
	"github.com/maksymenkoml/lingoose/observer/langfuse"
)

// download https://raw.githubusercontent.com/hwchase17/chat-your-data/master/state_of_the_union.txt

func main() {
	ctx := context.Background()

	o := langfuse.New(ctx)
	trace, err := o.Trace(&observer.Trace{Name: "state of the union"})
	if err != nil {
		panic(err)
	}

	ctx = observer.ContextWithObserverInstance(ctx, o)
	ctx = observer.ContextWithTraceID(ctx, trace.ID)

	qa := qa.New(
		openai.New().WithTemperature(0),
		index.New(
			jsondb.New().WithPersist("db.json"),
			openaiembedder.New(openaiembedder.AdaEmbeddingV2),
		),
	)

	_, err = os.Stat("db.json")
	if os.IsNotExist(err) {
		err = qa.AddSource(ctx, "state_of_the_union.txt")
		if err != nil {
			panic(err)
		}
	}

	response, err := qa.Run(ctx, "What is the NATO purpose?")
	if err != nil {
		panic(err)
	}

	fmt.Println(response)
}
