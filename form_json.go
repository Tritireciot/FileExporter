package main

import (
	"fmt"
	"maps"
	"os"
	"strings"

	"github.com/tidwall/gjson"
)


func getPair(dict map[string]string) (string, string) {
	for key, value := range dict {
		return key, value
	}
	return "", ""
}


func main() {
	data, err := os.ReadFile("rundown_print_2.json")
	if err != nil {
		fmt.Println("Pizdec", err.Error())
	}
	requiredTags := map[string]map[string]any{
		"RunDown" : {
			"Name": "rundown.title",
		},
		"Story" : {
			"Pos": "rundown.content.%d.pos",
			"DurationFact": "rundown.content.%d.story.fact_duration",
			"Type": "rundown.content.%d.story.type",
			"Name": "rundown.content.%d.story.name",
			"Media": map[string]string{
				"Duration": "rundown.content.%d.story.media_content.%d.duration",
			},
		},
	}
	repeatTags := map[string]string{
		"Story": "rundown.content",
	}

	resultData := map[string]any{}
	for tag := range repeatTags {
		resultData[tag] = []map[string]string{}
	}

	for tag, schema := range requiredTags{
		if list_json_path, ok := repeatTags[tag]; ok {
			list_schema, _ := resultData[tag].([]map[string]any)
			for i := range gjson.GetBytes(data, list_json_path).Array() {
				temp_schema := maps.Clone(schema)
				var toAdd bool = true
				for field, value := range schema {
					switch real_value := value.(type) {
					case string:
						res := gjson.GetBytes(data, fmt.Sprintf(real_value, i))
						if res.Exists(){
							temp_schema[field] = res.String()
						} else {
							toAdd = false
							break
						}
					case map[string]string:
						temp_res := []map[string]string{}
						_, json_path := getPair(real_value)
						list_json_subpath := json_path[:strings.LastIndex(json_path, ".%d")]
						subschema, ok := temp_schema[field].(map[string]string)
							if !ok {
								fmt.Println("No")
								break
							}
						for j := range gjson.GetBytes(data, fmt.Sprintf(list_json_subpath, i)).Array() {
							temp_subschema := maps.Clone(subschema)
							
							for sub_field, json_path := range real_value {
								res := gjson.GetBytes(data, fmt.Sprintf(json_path, i, j))
								temp_subschema[sub_field] = res.String()
							}

							temp_res = append(temp_res, temp_subschema)
						}
						temp_schema[field] = temp_res
					}
					
				}
				if toAdd {
					list_schema = append(list_schema, temp_schema)
				}

				
			}
			resultData[tag] = list_schema
		} else {
			
			for field, value := range schema {
				switch real_value := value.(type) {
				case string:
					res := gjson.GetBytes(data, real_value)
					schema[field] = res.String()
				}
			}
			resultData[tag] = schema
		}

	}




	//res := gjson.GetBytes(data, "rundown.content")
	fmt.Println(resultData)

}
