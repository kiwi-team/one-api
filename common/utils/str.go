package utils

func GetStringIfEmpty(str string, defaultValue string) string {
	if str == "" {
		return defaultValue
	}
	return str
}
