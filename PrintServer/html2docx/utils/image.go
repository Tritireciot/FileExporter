package utils

import (
	"encoding/base64"
	"image"
	"strings"
)

func GetImageDimensions(base64Str string) (int, int, error) {

	if idx := strings.Index(base64Str, ","); idx != -1 {
		base64Str = base64Str[idx+1:]
	}

	b64Reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(base64Str))
	config, _, err := image.DecodeConfig(b64Reader)
	if err != nil {
		return 0, 0, err
	}

	return config.Width, config.Height, nil
}
