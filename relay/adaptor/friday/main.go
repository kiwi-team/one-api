package friday

import (
	"fmt"

	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func GetRequestURL(meta *meta.Meta) (string, error) {
	switch meta.Mode {
	case relaymode.ChatCompletions:
		return fmt.Sprintf("%s/v1/openai/native/chat/completions", meta.BaseURL), nil
	case relaymode.ImagesGenerations:
		if meta.OriginModelName == "LongCat-T2I-Medium" {
			return fmt.Sprintf("%s/v1/openai/native/chat/completions", meta.BaseURL), nil
		}
		return fmt.Sprintf("%s/v1/openai/native/images/generations", meta.BaseURL), nil
	case relaymode.Embeddings:
		return fmt.Sprintf("%s/v1/openai/native/embeddings", meta.BaseURL), nil
	default:
	}
	return "", fmt.Errorf("unsupported relay mode %d for friday", meta.Mode)
}
