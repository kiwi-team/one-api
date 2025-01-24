package channeltype

import (
	"github.com/songquanpeng/one-api/relay/apitype"
)

func ToAPIType(channelType int) int {
	apiType := apitype.OpenAI
	switch channelType {
	case Anthropic:
		apiType = apitype.Anthropic
	case Baidu:
		apiType = apitype.Baidu
	case PaLM:
		apiType = apitype.PaLM
	case Zhipu:
		apiType = apitype.Zhipu
	case Ali:
		apiType = apitype.Ali
	case Xunfei:
		apiType = apitype.Xunfei
	case AIProxyLibrary:
		apiType = apitype.AIProxyLibrary
	case Tencent:
		apiType = apitype.Tencent
	case Gemini:
		apiType = apitype.Gemini
	case Ollama:
		apiType = apitype.Ollama
	case AwsClaude:
		apiType = apitype.AwsClaude
	case Coze:
		apiType = apitype.Coze
	case Cohere:
		apiType = apitype.Cohere
	case Cloudflare:
		apiType = apitype.Cloudflare
	case DeepL:
		apiType = apitype.DeepL
	case VertextAI:
		apiType = apitype.VertexAI
	case Proxy:
		apiType = apitype.Proxy
	case ImaginePro:
		apiType = apitype.ImaginePro
	case KlingAI:
		apiType = apitype.KlingAI
	case BFL:
		apiType = apitype.BFL
	case AI302:
		apiType = apitype.AI302
	case Midjourney:
		apiType = apitype.Midjourney
	}

	return apiType
}
