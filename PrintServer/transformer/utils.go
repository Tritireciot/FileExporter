package transformer

import (
	"AutoplayX/paths"
	"PrintServer/PrintServer/config"
	logging "PrintServer/agent"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

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

	logging.Agent.AddSimpleInfo("Получение шаблона", "printcontent: "+string(body)+resp.Status+"\n"+req.URL.Path)

	err = json.Unmarshal(body, template)
	return err

}

func getPrintData(context any, template *Template) (map[string]any, error) {
	client := &http.Client{}
	data := map[string]any{
		"context": context,
		"type":    template.Type,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	printurl := config.GetSubAccess().GetPath()
	if config.PrintContentHost != "" && config.PrintContentPort != "" {
		printurl = config.PrintContentHost + ":" + config.PrintContentPort
	}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"http://%s/api/%s/svc/printcontent",
			printurl,
			template.Subsystem,
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

	var printData map[string]any
	logging.Agent.AddSimpleInfo("Получение данных для печати", "printcontent: "+string(body)+resp.Status+"\n"+req.URL.Path + "\n" + printurl)

	err = json.Unmarshal(body, &printData)

	return printData, err
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

func FramesToTime(frames int) string {
	fps := 25
	rest_frames := frames % fps
	seconds := frames / fps
	rest_seconds := seconds % 60
	minutes := seconds / 60
	rest_minutes := minutes % 60
	hours := minutes / 60
	return fmt.Sprintf("%02d:%02d:%02d:%02d", hours, rest_minutes, rest_seconds, rest_frames)
}
