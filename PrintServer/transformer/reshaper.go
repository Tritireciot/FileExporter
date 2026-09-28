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

func (reshaper *Reshaper) ChangeBaseTags(template_content string, repeats []string) string {
	return reshaper.tagPattern.ReplaceAllStringFunc(template_content, func(tag string) string {
		match_tag := reshaper.tagPattern.FindStringSubmatch(tag)
		entity, field, funcName := match_tag[1], match_tag[2], match_tag[3]

		if funcName != "" {
			field += " | " + funcName
		}
		if contains(repeats, entity) {
			return fmt.Sprintf("{{ $%s.%s }}", entity, field)
		}
		return fmt.Sprintf("{{ .%s.%s }}", entity, field)
	})
}

func (reshaper *Reshaper) ChangeRepeatTags(template_content string, render_data map[string]bool) ([]string, string) {
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

func (reshaper *Reshaper) TransformTemplate(template Template) (string, error) {
	template_content := template.Content
	template_content = "{{ $hasContent := false }}" + template_content
	repeats, template_content := reshaper.ChangeRepeatTags(template_content, template.RenderData)
	logging.Agent.AddSimpleInfo("Подготовка шаблона печати", "Заменены теги повторов: "+fmt.Sprint(repeats))
	template_content = reshaper.ChangeBaseTags(template_content, repeats)

	return template_content, nil
}
