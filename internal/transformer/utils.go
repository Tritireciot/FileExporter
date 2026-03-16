package transformer

import (
	"context"
	"former/internal/db"
)

func collectAliases(ctx context.Context, db_repo *db.DBRepository, requiredTags *map[string]string) error {
	for tag_name := range *requiredTags {
		tag := db.Tag{Name: tag_name}
		err := db_repo.GetElementByName(ctx, &tag)
		if err != nil {
			return err
		}
		(*requiredTags)[tag_name] = tag.Alias
	}
	return nil

}