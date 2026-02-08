package tools

import "encoding/json"

func CopyMapViaJSON[T any](src T) (T, error) {
	var dst T
	b, err := json.Marshal(src)
	if err != nil {
		return dst, err
	}
	err = json.Unmarshal(b, &dst)
	return dst, err
}
