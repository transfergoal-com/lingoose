package llm_with_usage

import (
	"context"
	"time"

	"github.com/maksymenkoml/lingoose/thread"
)

const (
	defaultTimeoutInMinutes = 6
)

type TokensUsage struct {
	PromptTokens     int
	CompletionTokens int
	AudioTokens      int
	CachedTokens     int
}

type LLMWithUsage interface {
	GenerateWithUsage(context.Context, *thread.Thread) (*TokensUsage, error)
}

type Tool struct {
	llm LLMWithUsage
}

func New(llm LLMWithUsage) *Tool {
	return &Tool{
		llm: llm,
	}
}

type Input struct {
	Query string `json:"query" jsonschema:"description=user query"`
}

type Output struct {
	Error  string       `json:"error,omitempty"`
	Result string       `json:"result,omitempty"`
	Usage  *TokensUsage `json:"usage,omitempty"`
}

type FnPrototype func(Input) Output

func (t *Tool) Name() string {
	return "llm_with_usage"
}

func (t *Tool) Description() string {
	return "A tool that uses a language model to generate a response to a user query and returns token usage information."
}

func (t *Tool) Fn() any {
	return t.fn
}

//nolint:gosec
func (t *Tool) fn(i Input) Output {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeoutInMinutes*time.Minute)
	defer cancel()

	th := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent(i.Query),
		),
	)

	usage, err := t.llm.GenerateWithUsage(ctx, th)
	if err != nil {
		return Output{Error: err.Error()}
	}

	return Output{
		Result: th.LastMessage().Contents[0].AsString(),
		Usage:  usage,
	}
}
