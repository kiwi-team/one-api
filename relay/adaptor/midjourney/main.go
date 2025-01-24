package midjourney

import (
	"fmt"

	"github.com/songquanpeng/one-api/relay/model"
)

func ConvertImageRequest(request model.ImageRequest) *MJImageSubmitTaskRequest {
	var imageRequest MJImageSubmitTaskRequest
	imageRequest.Model = request.Model
	imageRequest.Prompt = fmt.Sprintf(`%s %s`, request.Prompt, " --v 6.1")
	fmt.Printf("%v", imageRequest)
	return &imageRequest
}
