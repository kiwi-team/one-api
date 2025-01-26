package baidu

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Baidu-AIP/golang-sdk/aip/censor"
	"github.com/songquanpeng/one-api/common/config"
)

type Adaptor struct {
	Source string
}

/*
	{
	    "log_id": 15572142621780024,
	    "conclusion": "合规",
	    "conclusionType": 1,
	    "data": [{
	        "type": 14,
	        "subType": 0,
	        "conclusion": "合规",
	        "conclusionType": 1,
	        "msg": "自定义文本白名单审核通过",
	        "hits": [{
	            "datasetName": "SLK-测试-自定义文本白名单",
	            "words": ["习大大"]
	        }]
	    }]
	}
*/
type Hit struct {
	DatasetName string   `json:"datasetName"`
	Words       []string `json:"words"`
}

type Data struct {
	Type           int    `json:"type"`
	SubType        int    `json:"subType"`
	Conclusion     string `json:"conclusion"`
	ConclusionType int    `json:"conclusionType"`
	Msg            string `json:"msg"`
	Hits           []Hit  `json:"hits"`
}

type BaiduTextCensorResult struct {
	LogID          int    `json:"log_id"`
	Conclusion     string `json:"conclusion"`
	ConclusionType int    `json:"conclusionType"`
	Data           []Data `json:"data"`
}

type ErrorResponse struct {
	LogID     int    `json:"log_id"`
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

var client *censor.ContentCensorClient

func init() {
	fmt.Println("baidu content check")
}

// https://cloud.baidu.com/doc/ANTIPORN/s/2kvuvd2pr#%E8%BF%94%E5%9B%9E%E5%8F%82%E6%95%B0%E8%AF%A6%E6%83%85-1
func (a *Adaptor) CheckContent(content string, channel_id int) (bool, string, error) {
	parts := strings.Split(config.BaiduAKAndSK, ";")
	client = censor.NewClient(parts[0], parts[1])
	ret := client.TextCensor(content)
	return ParseResponse(ret)
}

func ParseResponse(jsonStr string) (bool, string, error) {
	// 尝试解析为 BaiduTextCensorResult
	var result BaiduTextCensorResult
	fmt.Println("百度接口返回：" + jsonStr)
	if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
		return result.ConclusionType == 1, jsonStr, nil
	}
	return true, jsonStr, fmt.Errorf("baidu文本审查接口返回异常")
}
