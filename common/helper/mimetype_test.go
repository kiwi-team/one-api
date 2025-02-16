package helper

import (
	"fmt"
	"testing"
)

func TestDetectFileFromURL(t *testing.T) {
	url := "https://llmstatic.s3.cn-northwest-1.amazonaws.com.cn/e568dd8dce441eecc168bde02c261d50a/3d100202-6ff6-4191-9f6a-ee191f9bfce3_0.png"
	mime, err := DetectFileFromURL(url)
	fmt.Println(mime, err)

	url = "https://llmstatic.s3.cn-northwest-1.amazonaws.com.cn/e568dd8dce441eecc168bde02c261d50a/file-1739520873776xuMHkFvpj9xc.png"
	mime, err = DetectFileFromURL(url)
	fmt.Println(mime, err)
}

func TestDetectFileAndBase64File(t *testing.T) {
	url := "https://llmstatic.s3.cn-northwest-1.amazonaws.com.cn/e568dd8dce441eecc168bde02c261d50a/file-1739520908853efmpETW3eCfP.wav"
	mime, base64Data, err := DetectFileAndBase64File(url)
	fmt.Println(mime, base64Data, err)
}
