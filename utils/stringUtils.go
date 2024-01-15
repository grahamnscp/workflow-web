package utils

import (
	"encoding/base64"
)

/* base64 decode string */
func DecodeB64(message string) (retour string) {
	base64Text := make([]byte, base64.StdEncoding.DecodedLen(len(message)))
	base64.StdEncoding.Decode(base64Text, []byte(message))
	return string(base64Text)
}

/* Strip any nonalpha chars */
func ClearString(str string) string {
	return nonAlphanumericRegex.ReplaceAllString(str, "")
}

/* Remove nil fields in map[string]interface{} */
func RemoveNils(initialMap map[string]interface{}) map[string]interface{} {
	withoutNils := map[string]interface{}{}
	for key, value := range initialMap {
		_, ok := value.(map[string]interface{})
		if ok {
			value = RemoveNils(value.(map[string]interface{}))
			withoutNils[key] = value
			continue
		}
		if value != nil {
			withoutNils[key] = value
		}
	}
	return withoutNils
}
