package midjourney

type MJImageSubmitTaskRequest struct {
	Model  string `json:"model,omitempty"`
	Prompt string `json:"prompt"`
}

type MJImageSubmitTaskResponse struct {
	Code        int64  `json:"code"`
	Description string `json:"description"`
	Properties  struct {
		DiscordChannelID  string `json:"discordChannelId"`
		DiscordInstanceID string `json:"discordInstanceId"`
	} `json:"properties"`
	Result string `json:"result"`
}

type MJImageQueryTaskResponse struct {
	Action  string `json:"action"`
	BotType string `json:"botType"`
	Buttons []struct {
		CustomID string `json:"customId"`
		Emoji    string `json:"emoji"`
		Label    string `json:"label"`
		Style    int64  `json:"style"`
		Type     int64  `json:"type"`
	} `json:"buttons"`
	CustomID    string `json:"customId"`
	Description string `json:"description"`
	FailReason  string `json:"failReason"`
	FinishTime  int64  `json:"finishTime"`
	ID          string `json:"id"`
	ImageURL    string `json:"imageUrl"`
	MaskBase64  string `json:"maskBase64"`
	Progress    string `json:"progress"`
	Prompt      string `json:"prompt"`
	PromptEn    string `json:"promptEn"`
	Properties  struct {
		FinalPrompt   string `json:"finalPrompt"`
		FinalZhPrompt string `json:"finalZhPrompt"`
	} `json:"properties"`
	StartTime  int64  `json:"startTime"`
	State      string `json:"state"`
	Status     string `json:"status"`
	SubmitTime int64  `json:"submitTime"`
}
