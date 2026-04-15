package transformer

import (
	"context"
	"fmt"
	"former/internal/db"
	"regexp"
)

type Reshaper struct {
	db_repo          db.DBRepo
	tagPattern       *regexp.Regexp
	tagRepeatPattern [2]*regexp.Regexp
}

func NewReshaper(db_repo db.DBRepo) *Reshaper {
	return &Reshaper{
		db_repo:    db_repo,
		tagPattern: regexp.MustCompile(`#([A-Za-z]+)\.([A-Za-z]+)#`),
		tagRepeatPattern: [2]*regexp.Regexp{
			regexp.MustCompile(`<#Repeat#([A-Za-z]+)#>`),
			regexp.MustCompile(`</#Repeat#([A-Za-z]+)#>`),
		},
	}
}

func contains(slice []string, element string) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

func (reshaper *Reshaper) ChangeBaseTags(ctx context.Context, template_content string, repeats []string, requiredTags *map[string]any) string {
	return reshaper.tagPattern.ReplaceAllStringFunc(template_content, func(tag string) string {
		match_tag := reshaper.tagPattern.FindStringSubmatch(tag)
		entity, field := match_tag[1], match_tag[2]
		if sub_map, ok := (*requiredTags)[entity]; ok {
			switch typed_sub_map := sub_map.(type) {
			case map[string]any:
				typed_sub_map[field] = getAlias(ctx, reshaper.db_repo, match_tag[0])
				sub_map = typed_sub_map
			}
			(*requiredTags)[entity] = sub_map
		} else {
			(*requiredTags)[entity] = map[string]any{field: getAlias(ctx, reshaper.db_repo, match_tag[0])}
		}
		if contains(repeats, entity) {
			return fmt.Sprintf("{{ $%s.%s }}", entity, field)
		}
		return fmt.Sprintf("{{ .%s.%s }}", entity, field)
	})
}

func (reshaper *Reshaper) ChangeRepeatTags(ctx context.Context, template_content string, repeatTags *map[string]string) ([]string, string) {
	repeats := []string{}
	template_content = reshaper.tagRepeatPattern[0].ReplaceAllStringFunc(template_content, func(tag string) string {
		match_tag := reshaper.tagRepeatPattern[0].FindStringSubmatch(tag)
		field := match_tag[1]
		repeats = append(repeats, field)
		(*repeatTags)[field] = getAlias(ctx, reshaper.db_repo, match_tag[0][1:len(match_tag[0])-1])
		return fmt.Sprintf("{{ range $%s := .%s }}", field, field)
	})

	return repeats, reshaper.tagRepeatPattern[1].ReplaceAllStringFunc(template_content, func(tag string) string {
		return "{{ end }}"
	})

}

func (reshaper *Reshaper) TransformTemplate(ctx context.Context, template_content string, requiredTags *map[string]any, repeatTags *map[string]string) string {
	connections := IncludeRepeatStructure(template_content, requiredTags)
	repeats, template_content := reshaper.ChangeRepeatTags(ctx, template_content, repeatTags)
	template_content = reshaper.ChangeBaseTags(ctx, template_content, repeats, requiredTags)
	CompleteConnections(connections, requiredTags, repeatTags)
	return template_content
}
