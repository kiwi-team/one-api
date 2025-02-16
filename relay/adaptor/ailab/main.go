package ailab

import (
	"github.com/songquanpeng/one-api/relay/model"
)

func ConvertImageRequest(request model.ImageRequest) *ImageSubmitTaskRequest {
	var imageRequest ImageSubmitTaskRequest
	imageRequest.Model = request.Model
	imageRequest.Prompt = request.Prompt
	imageRequest.Size = "1:1"
	imageRequest.Type = 1
	return &imageRequest
}
