package converters

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"io"
	"strings"

	"github.com/mmonterroca/docxgo/v2/domain"
)

type Image struct {
	Size   domain.ImageSize
	Data   []byte
	Format domain.ImageFormat
}

var formats = map[string]domain.ImageFormat{
	"png":  domain.ImageFormatPNG,
	"jpeg": domain.ImageFormatJPEG,
	"gif":  domain.ImageFormatGIF,
	"webp": domain.ImageFormatWEBP,
	"svg":  domain.ImageFormatSVG,
	"tiff": domain.ImageFormatTIFF,
	"bmp":  domain.ImageFormatBMP,
}

func ConvertImage(base64Str string) (*Image, error) {
	basePrefix := "data:image/"
	if !strings.HasPrefix(base64Str, basePrefix) {
		return nil, fmt.Errorf("Не поддерживаемый формат. Только base64")
	}

	baseMarker := ";base64,"

	format_end := strings.Index(base64Str, baseMarker)
	if format_end == -1 {
		return nil, fmt.Errorf("Маркер base64 не найден")
	}
	mimeType, ok := formats[base64Str[len(basePrefix):format_end]]
	if !ok {
		return nil, fmt.Errorf("Не поддерживаемый формат.")
	}

	base64RawString := base64Str[format_end+len(baseMarker):]

	stringReader := strings.NewReader(base64RawString)
	base64Reader := base64.NewDecoder(base64.StdEncoding, stringReader)

	var buffer bytes.Buffer
	teeReader := io.TeeReader(base64Reader, &buffer)

	cfg, _, err := image.DecodeConfig(teeReader)
	if err != nil {
		return nil, fmt.Errorf("не удалось определить размеры картинки: %w", err)
	}

	_, err = io.Copy(&buffer, base64Reader)
	if err != nil {
		return nil, fmt.Errorf("ошибка полного декодирования base64: %w", err)
	}

	return &Image{
		Size:   domain.NewImageSize(cfg.Width, cfg.Height),
		Format: mimeType,
		Data:   buffer.Bytes(),
	}, nil

}
