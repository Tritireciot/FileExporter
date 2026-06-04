package db

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func formatPathName(path string, separator string) string {
	path = strings.TrimSuffix(path, filepath.Ext(path))
	return strings.Join(strings.Split(path, "/")[2:], separator)
}

func (repository *DBRepository) fillTemplatesTable(ctx context.Context, templates_path string) error {

	err := filepath.WalkDir(templates_path, func(path string, d fs.DirEntry, err error) error {

		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".html" {
			return nil
		}

		data, err := os.ReadFile(path)

		if err != nil {
			return err
		}
		template_name := formatPathName(path, ".")
		repository.logger.Println("Insert Template: ", template_name)
		if err := repository.AddTemplate(ctx, &Template{Name: template_name, Content: string(data), Subsystem: strings.ToLower(strings.Split(path, "/")[1]), IsActive: true}); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (repository *DBRepository) fillTagsTable(ctx context.Context, tags_filepath string) error {
	var tags []Tag

	file, err := os.Open(tags_filepath)

	if err != nil {
		return err
	}

	if err := json.NewDecoder(file).Decode(&tags); err != nil {
		return err
	}
	for _, tag_data := range tags {
		repository.logger.Println("Insert: ", tag_data.Name)
		if err := repository.AddTag(ctx, &tag_data); err != nil {
			return err
		}
	}

	return nil
}
