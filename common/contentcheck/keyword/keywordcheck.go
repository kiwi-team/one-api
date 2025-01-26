package keyword

import (
	"fmt"
	"strings"

	"github.com/songquanpeng/one-api/common/config"
)

type Adaptor struct{}

var keywords []string

func init() {
	keywords = strings.Split(strings.ToLower(config.ContentCheckKeyWords), "\n")
	fmt.Println("keyword content check")
}
func (a *Adaptor) CheckContent(content string, channel_id int) (bool, string, error) {
	input := strings.ToLower(content)
	for _, keyword := range keywords {
		if strings.Contains(input, keyword) {
			return true, fmt.Sprintf("keyword:%s", keyword), nil
		}
	}
	return false, "", nil
}
