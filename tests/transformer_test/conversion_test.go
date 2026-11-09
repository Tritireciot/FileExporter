package transformer_test

import (
	"PrintServer/PrintServer/transformer"
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestFormatters(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"date", transformer.DashedYMDtoPrintDate("2026-11-09"), "9 Ноября 2026"},
		{"invalid date", transformer.DashedYMDtoPrintDate("invalid"), "invalid"},
		{"weekday", transformer.ShortWeekday("09.11.2026"), "9 Пн"},
		{"invalid weekday", transformer.ShortWeekday("invalid"), "invalid"},
		{"frames", transformer.FramesToTime(25*3661 + 12), "01:01:01:12"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestDOCXFromTemplateContainsText(t *testing.T) {
	service := &transformer.TransformService{}
	data, err := service.DOCXFromTemplate("<html><body><p>Transformer document</p></body></html>")
	if err != nil {
		t.Fatal(err)
	}
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("DOCX is not a ZIP archive: %v", err)
	}
	for _, file := range archive.File {
		if file.Name != "word/document.xml" {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), "Transformer document") {
			t.Fatalf("generated document lost paragraph text: %s", body)
		}
		return
	}
	t.Fatal("DOCX has no word/document.xml")
}
