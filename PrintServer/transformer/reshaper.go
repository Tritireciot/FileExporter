package transformer

import (
	logging "PrintServer/agent"
	"fmt"
	"regexp"
	"strings"
)

type Reshaper struct {
	tagPattern       *regexp.Regexp
	tagRepeatPattern [4]*regexp.Regexp
}

func NewReshaper() *Reshaper {
	return &Reshaper{
		tagPattern: regexp.MustCompile(`#([A-Za-z]+)\.([A-Za-z]+)(?:\s*\|\s*([A-Za-z]+))?#`),
		tagRepeatPattern: [4]*regexp.Regexp{
			regexp.MustCompile(`<#Repeat#([A-Za-z]+)#>`),
			regexp.MustCompile(`</#Repeat#([A-Za-z]+)#>`),
			regexp.MustCompile(`<#At#([A-Za-z0-9_]+)\[([0-9]+)\]#>`),
			regexp.MustCompile(`</#At#([A-Za-z]+)#>`),
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

func (reshaper *Reshaper) ChangeBaseTags(template_content string, repeats []string, requiredTags *map[string]any, tag_list *[]string) string {
	return reshaper.tagPattern.ReplaceAllStringFunc(template_content, func(tag string) string {
		match_tag := reshaper.tagPattern.FindStringSubmatch(tag)
		entity, field, funcName := match_tag[1], match_tag[2], match_tag[3]
		if sub_map, ok := (*requiredTags)[entity]; ok {
			switch typed_sub_map := sub_map.(type) {
			case map[string]any:
				tag_name := fmt.Sprintf("#%s.%s#", entity, field)
				*tag_list = append(*tag_list, tag_name)
				typed_sub_map[field] = tag_name
				sub_map = typed_sub_map
			}
			(*requiredTags)[entity] = sub_map
		} else {
			tag_name := fmt.Sprintf("#%s.%s#", entity, field)
			*tag_list = append(*tag_list, tag_name)
			(*requiredTags)[entity] = map[string]any{field: tag_name}
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

func (reshaper *Reshaper) ChangeRepeatTags(template_content string, repeatTags *map[string]string, render_data map[string]bool, tag_list *[]string) ([]string, string) {
	repeats := []string{}

	template_content = reshaper.tagRepeatPattern[2].ReplaceAllStringFunc(template_content, func(tag string) string {
		match_tag := reshaper.tagRepeatPattern[2].FindStringSubmatch(tag)
		field := match_tag[1]
		pos := match_tag[2]

		return fmt.Sprintf("{{ with $%s := index .%s %s }}", field, field, pos)
	})

	template_content = reshaper.tagRepeatPattern[0].ReplaceAllStringFunc(template_content, func(tag string) string {
		match_tag := reshaper.tagRepeatPattern[0].FindStringSubmatch(tag)
		field := match_tag[1]

		res := []string{}
		repeats = append(repeats, field)
		tag_name := match_tag[0][1:len(match_tag[0])-1]
		*tag_list = append(*tag_list, tag_name)
		(*repeatTags)[field] = tag_name
		res = append(res, fmt.Sprintf("{{ range $%s := .%s }}", field, field))

		element := strings.ToLower(match_tag[1]) + "_break"
		if addBreak, ok := render_data[element]; ok && addBreak {
			res = append(res, "{{if $hasContent}}\n\t<div style=\"page-break-before: always;\"></div>\n{{end}}{{ $hasContent = true }}\n")
		}

		return strings.Join(res, "\n")
	})

	template_content = reshaper.tagRepeatPattern[1].ReplaceAllStringFunc(template_content, func(tag string) string {
		return "{{ end }}"
	})

	template_content = reshaper.tagRepeatPattern[3].ReplaceAllStringFunc(template_content, func(tag string) string {
		return "{{ end }}"
	})

	return repeats, template_content

}

func (reshaper *Reshaper) TransformTemplate(template Template, requiredTags *map[string]any, repeatTags *map[string]string) (string, error) {
	template_content := template.Content
	connections := IncludeRepeatStructure(template_content, requiredTags)
	logging.Agent.AddSimpleInfo("Подготовка шаблона печати", "Сформированы связи повторяющихся тегов: "+fmt.Sprint(connections))
	template_content = "{{ $hasContent := false }}" + template_content
	tag_list := []string{}
	repeats, template_content := reshaper.ChangeRepeatTags(template_content, repeatTags, template.RenderData, &tag_list)
	logging.Agent.AddSimpleInfo("Подготовка шаблона печати", "Заменены теги повторов: "+fmt.Sprint(repeats))
	template_content = reshaper.ChangeBaseTags(template_content, repeats, requiredTags, &tag_list)
	CompleteConnections(connections, requiredTags, repeatTags)

	aliases, err := getAliases(tag_list, template.Subsystem)
	if err != nil {
		return "", err
	}

	logging.Agent.AddSimpleInfo("Подготовка шаблона печати", fmt.Sprint(*requiredTags))
	logging.Agent.AddSimpleInfo("Подготовка шаблона печати", fmt.Sprint(*repeatTags))
	fillMap(requiredTags, aliases)
	for repKey, repValue := range *repeatTags {
		for _, alias := range aliases{
			if repValue == alias.Name {
				(*repeatTags)[repKey] = alias.Alias
			}
		}

	}
	logging.Agent.AddSimpleInfo("Подготовка шаблона печати", fmt.Sprint(*requiredTags))
	logging.Agent.AddSimpleInfo("Подготовка шаблона печати", fmt.Sprint(*repeatTags))


	logging.Agent.AddSimpleInfo("Подготовка шаблона печати", "Заменены теги: "+fmt.Sprint(requiredTags))
	return template_content, nil
}

func fillMap(someMap *map[string]any, data []TagAlias) {
	for key, value := range *someMap {
		if subdict, ok := value.(map[string]any); ok {
			fillMap(&subdict, data)
		} else {
			for _, tag := range data {
				if tag.Name == value {
					(*someMap)[key] = tag.Alias
				}
			}
		}
	}
}
