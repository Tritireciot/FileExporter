package transformer

import (
	"fmt"
	"maps" // TODO  убрать нет в 1.19
	"strings"

	"github.com/tidwall/gjson"
)

func getPair(dict map[string]string) (string, string) {
	for key, value := range dict {
		return key, value
	}
	return "", ""
}

func translate(json_data []byte, requiredTags map[string]map[string]any, repeatTags map[string]string) map[string]any {
	resultData := map[string]any{}
	for tag := range repeatTags {
		resultData[tag] = []map[string]string{}
	}

	for tag, schema := range requiredTags{
		if list_json_path, ok := repeatTags[tag]; ok {
			list_schema, _ := resultData[tag].([]map[string]any)
			for i := range gjson.GetBytes(json_data, list_json_path).Array() {
				temp_schema := maps.Clone(schema) // TODO нет в 1.19
				var toAdd bool = true
				for field, value := range schema {
					switch real_value := value.(type) {
					case string:
						res := gjson.GetBytes(json_data, fmt.Sprintf(real_value, i))
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
						for j := range gjson.GetBytes(json_data, fmt.Sprintf(list_json_subpath, i)).Array() {
							temp_subschema := maps.Clone(subschema) // TODO нет в 1.19
							
							for sub_field, json_path := range real_value {
								res := gjson.GetBytes(json_data, fmt.Sprintf(json_path, i, j))
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
					res := gjson.GetBytes(json_data, real_value)
					schema[field] = res.String()
				}
			}
			resultData[tag] = schema
		}
	}

	return resultData
}