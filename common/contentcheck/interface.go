package contentcheck

import (
	"github.com/songquanpeng/one-api/common/contentcheck/baidu"
	"github.com/songquanpeng/one-api/common/contentcheck/keyword"
	"github.com/songquanpeng/one-api/common/contentcheck/regex"
)

const (
	Baidu   = "baidu"
	Keyword = "keyword"
	Regex   = "regex"
)

type Adaptor interface {
	CheckContent(content string, channel_id int) (bool, string, error)
}

func GetAdaptor(apiType string) Adaptor {
	switch apiType {
	case Baidu:
		return &baidu.Adaptor{}
	case Keyword:
		return &keyword.Adaptor{}
	case Regex:
		return &regex.Adaptor{}
	}
	return nil
}
