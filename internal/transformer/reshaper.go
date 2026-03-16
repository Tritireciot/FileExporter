package transformer

import (
	"fmt"
	"regexp"
)

type Reshaper struct {
	tagPattern *regexp.Regexp
	tagRepeatPattern [2]*regexp.Regexp
}

func NewReshaper() *Reshaper {
	return &Reshaper{
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

func (reshaper *Reshaper) changeBaseTags(template_content string, repeats []string, requiredTags *map[string]string) string {
	return reshaper.tagPattern.ReplaceAllStringFunc(template_content, func(tag string) string {
		match_tag := reshaper.tagPattern.FindStringSubmatch(tag)
		(*requiredTags)[match_tag[0]] = ""
		entity, field := match_tag[1], match_tag[2]
		if contains(repeats, entity) {
			return fmt.Sprintf("{{ $%s.%s }}", entity, field)
		}
		return fmt.Sprintf("{{ .%s.%s }}", entity, field)
	})
}

func (reshaper *Reshaper) changeRepeatTags(template_content string, repeats []string) ([]string, string) {
	
	template_content = reshaper.tagRepeatPattern[0].ReplaceAllStringFunc(template_content, func(tag string) string {
		match_tag := reshaper.tagRepeatPattern[0].FindStringSubmatch(tag)
		field := match_tag[1]
		repeats = append(repeats, field)
		return fmt.Sprintf("{{ range $%s := .%s }}", field, field)
	})

	return repeats, reshaper.tagRepeatPattern[1].ReplaceAllStringFunc(template_content, func(tag string) string {
		return "{{ end }}"
	})

}

func (reshaper *Reshaper) TransformTemplate(template_content string, requiredTags *map[string]string) string {
	var repeats []string
	repeats, template_content = reshaper.changeRepeatTags(template_content, repeats)
	template_content = reshaper.changeBaseTags(template_content, repeats, requiredTags)
	return template_content
}