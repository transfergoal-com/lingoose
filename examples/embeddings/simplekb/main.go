package main

import (
	"context"

	openaiembedder "github.com/maksymenkoml/lingoose/embedder/openai"
	"github.com/maksymenkoml/lingoose/index"
	"github.com/maksymenkoml/lingoose/index/option"
	"github.com/maksymenkoml/lingoose/index/vectordb/jsondb"
	qapipeline "github.com/maksymenkoml/lingoose/legacy/pipeline/qa"
	"github.com/maksymenkoml/lingoose/llm/openai"
	"github.com/maksymenkoml/lingoose/loader"
	"github.com/maksymenkoml/lingoose/textsplitter"
)

func main() {
	docs, _ := loader.NewPDFToTextLoader("./kb").WithTextSplitter(textsplitter.NewRecursiveCharacterTextSplitter(2000, 200)).Load(context.Background())
	index := index.New(jsondb.New(), openaiembedder.New(openaiembedder.AdaEmbeddingV2)).WithIncludeContents(true)
	index.LoadFromDocuments(context.Background(), docs)
	qapipeline.New(openai.NewChat().WithVerbose(true)).WithIndex(index).Query(context.Background(), "What is the NATO purpose?", option.WithTopK(1))
}
