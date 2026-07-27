package transformer

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed standalone-linux-64.zip
var weasyprintZipBytes []byte

func (service *TransformService) PDFFromTemplate(htmlContent string) ([]byte, error) {
	cmd := exec.Command(service.pdfExec, "-", "-")

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("ошибка создания StdoutPipe: %v", err)
	}

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("ошибка создания StdinPipe: %v", err)
	}

	cmd.Stderr = os.Stderr 

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("ошибка старта процесса: %v", err)
	}

	go func() {
		defer stdinPipe.Close()
		_, _ = io.WriteString(stdinPipe, htmlContent)
	}()

	var pdfBuffer bytes.Buffer
	_, err = io.Copy(&pdfBuffer, stdoutPipe)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения PDF из потока: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("процесс WeasyPrint завершился с ошибкой: %v", err)
	}

	return pdfBuffer.Bytes(), nil
}

func unzipBytes(zipBytes []byte, targetDir string) error {
	bodyReader := bytes.NewReader(zipBytes)
	zipReader, err := zip.NewReader(bodyReader, int64(len(zipBytes)))
	if err != nil {
		return err
	}

	for _, file := range zipReader.File {
		path := filepath.Join(targetDir, file.Name)
		
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		inFile, err := file.Open()
		if err != nil {
			f.Close()
			return err
		}

		_, err = io.Copy(f, inFile)
		inFile.Close()
		f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
