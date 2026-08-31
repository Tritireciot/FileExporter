package transformer

import (
	"AutoplayX/paths"
	"PrintServer/PrintServer/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type TagAlias struct {
	Name string `json:"name"`
	Alias string `json:"alias"`
}

func getAliases(tag_names []string, subsystem string) ([]TagAlias, error) {
	client := &http.Client{}
	jsonData, err := json.Marshal(tag_names)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(
		"POST", 
		fmt.Sprintf(
			"http://%s/api/%s/config/svc/print/db/tags_aliases", 
			config.GetSubAccess().GetPath(), 
			subsystem,
		), 
		bytes.NewReader(jsonData),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set(paths.SubAccessTokenTag, config.SubAccessCommKey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var aliases []TagAlias

	err = json.Unmarshal(body, &aliases)
	return aliases, err
}

func getTemplate(template *Template) error {
	client := &http.Client{}
	req, err := http.NewRequest(
		"GET", 
		fmt.Sprintf(
			"http://%s/api/%s/config/svc/print/db/get_template?id=%d", 
			config.GetSubAccess().GetPath(), 
			template.Subsystem,
			template.ID,
		), 
		nil,
	)
	if err != nil {
		return err
	}
	req.Header.Set(paths.SubAccessTokenTag, config.SubAccessCommKey)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, template)
	return err
	
}

func modifyStructure(entity_stack []string, requiredTags *map[string]any, connections *map[string][]string) {
	sub_requiredTags := requiredTags
	for i, entity := range entity_stack {
		sub_map, ok := (*sub_requiredTags)[entity].(map[string]any)
		if !ok {
			if i > 0 {
				if _, ok := (*connections)[entity_stack[i-1]]; !ok {
					(*connections)[entity_stack[i-1]] = []string{}
				}
				(*connections)[entity_stack[i-1]] = append((*connections)[entity_stack[i-1]], entity)
			}
			(*sub_requiredTags)[entity] = map[string]any{}
			sub_map, _ = (*sub_requiredTags)[entity].(map[string]any)
		}
		sub_requiredTags = &sub_map
	}
}

func IncludeRepeatStructure(template_content string, requiredTags *map[string]any) map[string][]string {
	connections := map[string][]string{}
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

func CompleteConnections(connections map[string][]string, requiredTags *map[string]any, repeatTags *map[string]string) {
	for source, destinations := range connections {
		for _, destination := range destinations {
			destination_map := (*requiredTags)[destination]
			if source_map, ok := (*requiredTags)[source].(map[string]any); ok {
				source_map[destination] = destination_map
				delete(*requiredTags, destination)
			}
			delete(*repeatTags, destination)
		}

	}
}

var ruMonths = [...]string{
	"", "Января", "Февраля", "Марта", "Апреля", "Мая", "Июня",
	"Июля", "Августа", "Сентября", "Октября", "Ноября", "Декабря",
}

var ruShortWeekDays = [...]string{
	"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб",
}

func DashedYMDtoPrintDate(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}

	return t.Format("2") + " " + ruMonths[t.Month()] + " " + t.Format("2006")

}

func ShortWeekday(dateStr string) string {
	t, err := time.Parse("02.01.2006", dateStr)
	if err != nil {
		return dateStr
	}

	return t.Format("2") + " " + ruShortWeekDays[t.Weekday()]

}
