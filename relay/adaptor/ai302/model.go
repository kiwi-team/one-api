package ai302

/*
{
	"images": [
		{
			"url": "https://file.302.ai/gpt/imgs/20241127/b3bb792266b744cc80c43d3b4e1b0b0c.webp",
			"content_type": "image/webp",
			"file_size": 292054
		}
	]
}
*/

type Recraft302ImageResponse struct {
	Images []struct {
		URL         string `json:"url"`
		ContentType string `json:"content_type"`
		FileSize    int    `json:"file_size"`
	} `json:"images"`
}

type Recraft302ImageRequest struct {
	Prompt    string `json:"prompt"`
	ImageSize struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"image_size,omitempty"`
	Style  string `json:"style,omitempty"`
	Colors []struct {
		R int `json:"r"`
		G int `json:"g"`
		B int `json:"b"`
	} `json:"colors,omitempty"`
}
