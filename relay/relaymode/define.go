package relaymode

const (
	Unknown = iota
	ChatCompletions
	Completions
	Embeddings
	Moderations
	ImagesGenerations
	Edits
	AudioSpeech
	AudioTranscription
	AudioTranslation
	Responses // https://platform.openai.com/docs/api-reference/responses/create
	// Proxy is a special relay mode for proxying requests to custom upstream
	Proxy
)
