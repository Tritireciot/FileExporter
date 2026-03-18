package transformer

import (
	"context"
	"fmt"
	"former/internal/db"
	"regexp"
	"strings"
)

func getAlias(ctx context.Context, db_repo *db.DBRepository, tag_name string) string {
	tag := db.Tag{Name: tag_name}
	err := db_repo.GetElementByName(ctx, &tag)
	if err != nil {
		fmt.Println(tag_name)
		fmt.Println("Error in getting alias: ", err.Error())
		return ""
	}
	return tag.Alias
}

func modifyStructure(entity_stack []string, requiredTags *map[string]any, connections *map[string]string) {
	sub_requiredTags := requiredTags
	for i, entity := range  entity_stack{
		sub_map, ok := (*sub_requiredTags)[entity].(map[string]any)
		if !ok {
			if i > 0 {
				(*connections)[entity_stack[i - 1]] = entity
			}
			(*sub_requiredTags)[entity] = map[string]any{}
			sub_map, _ = (*sub_requiredTags)[entity].(map[string]any)
		} 
		sub_requiredTags = &sub_map
	}
}

func includeRepeatStructure(template_content string, requiredTags *map[string]any) map[string]string {
	connections := map[string]string{}
	repeat_pattern := regexp.MustCompile(`</?#Repeat#([A-Za-z]+)#>`)
	stack := []string{}
	for _, tag := range repeat_pattern.FindAllStringSubmatch(template_content, -1) {
		real_tag, entity := tag[0], tag[1]
		if strings.Contains(real_tag, "/") {
			modifyStructure(stack, requiredTags, &connections)
			stack = stack[:len(stack)-1]
		} else {
			stack = append(stack, entity)
		}
	}
	return connections
}

func completeConnections(connections map[string]string, requiredTags *map[string]any, repeatTags *map[string]string) {
	for source, destination := range connections {
		destination_map := (*requiredTags)[destination]
		if source_map, ok := (*requiredTags)[source].(map[string]any); ok {
			source_map[destination] = destination_map
			delete(*requiredTags, destination)
		}
		delete(*repeatTags, destination)

	}
}