package i18n

import (
	"fmt"
	"strings"
)

const (
	English = "en"
	Chinese = "zh"
)

func Normalize(language string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(language))
	value = strings.ReplaceAll(value, "_", "-")
	switch value {
	case "", English, "eng", "english":
		return English, nil
	case Chinese, "cn", "zh-cn", "zh-hans", "chinese", "中文", "简体中文":
		return Chinese, nil
	default:
		return "", fmt.Errorf("unsupported language %q, expected en or zh", language)
	}
}

func IsChinese(language string) bool {
	normalized, err := Normalize(language)
	return err == nil && normalized == Chinese
}
