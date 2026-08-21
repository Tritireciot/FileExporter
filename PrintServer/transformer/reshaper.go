package transformer

import (
	"PrintServer/PrintServer/db"
	logging "PrintServer/agent"
	"context"
	"fmt"
	"regexp"
	"strings"
)

type Reshaper struct {
	db_repo          db.DBRepo
	tagPattern       *regexp.Regexp
	tagRepeatPattern [2]*regexp.Regexp
}

func NewReshaper(db_repo db.DBRepo) *Reshaper {
	return &Reshaper{
		db_repo:    db_repo,
		tagPattern: regexp.MustCompile(`#([A-Za-z]+)\.([A-Za-z]+)(?:\s*\|\s*([A-Za-z]+))?#`),
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
		entity, field, funcName := match_tag[1], match_tag[2], match_tag[3]
		if sub_map, ok := (*requiredTags)[entity]; ok {
			switch typed_sub_map := sub_map.(type) {
			case map[string]any:
				typed_sub_map[field] = getAlias(ctx, reshaper.db_repo, fmt.Sprintf("#%s.%s#", entity, field))
				sub_map = typed_sub_map
			}
			(*requiredTags)[entity] = sub_map
		} else {
			(*requiredTags)[entity] = map[string]any{field: getAlias(ctx, reshaper.db_repo, fmt.Sprintf("#%s.%s#", entity, field))}
		}
		if funcName != "" {
			field += " | " + funcName
		}
		if contains(repeats, entity) {
			return fmt.Sprintf("{{ $%s.%s }}", entity, field)
		}
		return fmt.Sprintf("{{ .%s.%s }}", entity, field)
	})
}

func (reshaper *Reshaper) ChangeRepeatTags(ctx context.Context, template_content string, repeatTags *map[string]string, render_data map[string]bool) ([]string, string) {
	repeats := []string{}
	template_content = reshaper.tagRepeatPattern[0].ReplaceAllStringFunc(template_content, func(tag string) string {
		match_tag := reshaper.tagRepeatPattern[0].FindStringSubmatch(tag)
		field := match_tag[1]

		res := []string{}
		repeats = append(repeats, field)
		(*repeatTags)[field] = getAlias(ctx, reshaper.db_repo, match_tag[0][1:len(match_tag[0])-1])
		res = append(res, fmt.Sprintf("{{ range $%s := .%s }}", field, field))

		element := strings.ToLower(match_tag[1]) + "_break"
		if addBreak, ok := render_data[element]; ok && addBreak {
			res = append(res, "{{if $hasContent}}\n\t<div style=\"page-break-before: always;\"></div>\n{{end}}{{ $hasContent = true }}\n")
		}

		return strings.Join(res, "\n")
	})

	return repeats, reshaper.tagRepeatPattern[1].ReplaceAllStringFunc(template_content, func(tag string) string {
		return "{{ end }}"
	})

}

func (reshaper *Reshaper) TransformTemplate(ctx context.Context, template_content string, requiredTags *map[string]any, repeatTags *map[string]string, render_data map[string]bool) string {
	connections := IncludeRepeatStructure(template_content, requiredTags)
	logging.Agent.AddSimpleInfo("Подготовка шаблона печати", "Сформированы связи повторяющихся тегов: "+fmt.Sprint(connections))
	template_content = "{{ $hasContent := false }}" + template_content
	repeats, template_content := reshaper.ChangeRepeatTags(ctx, template_content, repeatTags, render_data)
	logging.Agent.AddSimpleInfo("Подготовка шаблона печати", "Заменены теги повторов: "+fmt.Sprint(repeats))
	template_content = reshaper.ChangeBaseTags(ctx, template_content, repeats, requiredTags)
	CompleteConnections(connections, requiredTags, repeatTags)
	logging.Agent.AddSimpleInfo("Подготовка шаблона печати", "Заменены теги: "+fmt.Sprint(requiredTags))
	return template_content
}
