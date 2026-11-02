package transformer

import (
	"fmt"
	"time"
)

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
