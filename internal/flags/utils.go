package flags

import (
	"strings"

	"github.com/golang/glog"
)

// Decodes tags of the form "tag=key:value"
func DecodeTags(vals map[string][]string) map[string]string {
	if vals == nil {
		return nil
	}
	var tags map[string]string
	if len(vals["tag"]) > 0 {
		tags = make(map[string]string)
		tagList := vals["tag"]
		for _, tag := range tagList {
			s := strings.Split(tag, ":")
			if len(s) == 2 {
				k, v := s[0], s[1]
				tags[k] = v
			} else {
				glog.Warning("invalid tag ", tag)
			}
		}
	}
	return tags
}

func DecodeValue(vals map[string][]string, name string) string {
	value := ""
	if len(vals[name]) > 0 {
		value = vals[name][0]
	}
	return value
}