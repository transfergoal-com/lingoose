package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	openaiembedder "github.com/maksymenkoml/lingoose/embedder/openai"
	"github.com/maksymenkoml/lingoose/index"
	indexoption "github.com/maksymenkoml/lingoose/index/option"
	"github.com/maksymenkoml/lingoose/legacy/chat"

	"github.com/maksymenkoml/lingoose/index/vectordb/jsondb"
	"github.com/maksymenkoml/lingoose/legacy/prompt"
	"github.com/maksymenkoml/lingoose/llm/openai"
	"github.com/maksymenkoml/lingoose/loader"
	"github.com/maksymenkoml/lingoose/textsplitter"
	"github.com/maksymenkoml/lingoose/types"
)

const (
	KB = "https://en.wikipedia.org/wiki/World_War_II"
)

func main() {

	index := index.New(
		jsondb.New().WithPersist("db.json"),
		openaiembedder.New(openaiembedder.AdaEmbeddingV2),
	).WithIncludeContents(true)

	indexIsEmpty, _ := index.IsEmpty(context.Background())

	if indexIsEmpty {
		err := ingestData(index)
		if err != nil {
			panic(err)
		}
	}

	llmOpenAI := openai.NewChat()

	fmt.Println("Enter a query to search the knowledge base. Type 'quit' to exit.")
	query := ""
	for query != "quit" {

		fmt.Printf("> ")
		reader := bufio.NewReader(os.Stdin)
		query, _ := reader.ReadString('\n')

		if query == "quit" {
			break
		}

		similarities, err := index.Query(context.Background(), query, indexoption.WithTopK(3))
		if err != nil {
			panic(err)
		}

		content := ""

		for _, similarity := range similarities {
			fmt.Printf("Similarity: %f\n", similarity.Score)
			fmt.Printf("Document: %s\n", similarity.Content())
			fmt.Println("Metadata: ", similarity.Metadata)
			fmt.Println("----------")
			content += similarity.Content() + "\n"
		}

		systemPrompt := prompt.New("You are an helpful assistant. Answer to the questions using only " +
			"the provided context. Don't add any information that is not in the context. " +
			"If you don't know the answer, just say 'I don't know'.",
		)
		userPrompt := prompt.NewPromptTemplate(
			"Based on the following context answer to the question.\n\nContext:\n{{.context}}\n\nQuestion: {{.query}}").WithInputs(
			types.M{
				"query":   query,
				"context": content,
			},
		)

		chat := chat.New(
			chat.PromptMessage{
				Type:   chat.MessageTypeSystem,
				Prompt: systemPrompt,
			},
			chat.PromptMessage{
				Type:   chat.MessageTypeUser,
				Prompt: userPrompt,
			},
		)

		response, err := llmOpenAI.Chat(context.Background(), chat)
		if err != nil {
			panic(err)
		}

		fmt.Println(response)

	}

}

func ingestData(index *index.Index) error {

	fmt.Printf("Learning Knowledge Base...")

	loader := loader.NewPDFToTextLoader("./kb").WithPDFToTextPath("/opt/homebrew/bin/pdftotext")

	documents, err := loader.Load(context.Background())
	if err != nil {
		return err
	}

	textSplitter := textsplitter.NewRecursiveCharacterTextSplitter(2000, 200)

	documentChunks := textSplitter.SplitDocuments(documents)

	err = index.LoadFromDocuments(context.Background(), documentChunks)
	if err != nil {
		return err
	}

	fmt.Printf("Done\n")

	return nil
}
