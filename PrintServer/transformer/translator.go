package transformer

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

func getPair(dict map[string]string) (string, string) {
	for key, value := range dict {
		return key, value
	}
	return "", ""
}

func transformMap(input map[string]any) map[string]map[string]any {
	output := make(map[string]map[string]any)

	for key, value := range input {

		if nestedMap, ok := value.(map[string]any); ok {
			output[key] = nestedMap
		} else {
			fmt.Printf("Warning: value for key \"%s\" is not a map[string]any, it is %T\n", key, value)
		}
	}
	return output
}

func transformMapMedia(input map[string]any) map[string]string {
	output := make(map[string]string)

	for key, value := range input {

		if val, ok := value.(string); ok {
			output[key] = val
		} else {
			fmt.Printf("Warning: value for key \"%s\" is not a map[string]any, it is %T\n", key, value)
		}
	}
	return output
}

func mapClone[M ~map[K]V, K comparable, V any](dict M) M {
	new_dict := M{}
	for key, value := range dict {
		new_dict[key] = value
	}
	return new_dict
}

func addRepeatedTags(repeatTags map[string]string) map[string]any {
	resultData := map[string]any{}
	for tag := range repeatTags {
		resultData[tag] = []map[string]string{}
	}
	return resultData
}

func fillRepeatedTag(json_data []byte, list_json_path string, list_schema []map[string]any, schema map[string]any) []map[string]any {
	var elements []gjson.Result
	if list_json_path != "" {
		elements = gjson.GetBytes(json_data, list_json_path).Array()
	} else {
		elements = gjson.ParseBytes(json_data).Array()
	}
	for i := range elements {
		temp_schema := mapClone(schema)
		var toAdd bool = true
		for field, value := range schema {
			switch real_value := value.(type) {
			case string:
				res := gjson.GetBytes(json_data, fmt.Sprintf(real_value, i))
				if res.Exists() {
					temp_schema[field] = res.String()
				} else {
					toAdd = false
					break
				}
			case map[string]any:
				temp_res := []map[string]any{}
				media_map := transformMapMedia(real_value)
				_, json_path := getPair(media_map)
				list_json_subpath := json_path[:strings.LastIndex(json_path, ".%d")]
				subschema, ok := temp_schema[field].(map[string]any)
				if !ok {
					break
				}
				for j := range gjson.GetBytes(json_data, fmt.Sprintf(list_json_subpath, i)).Array() {
					temp_subschema := mapClone(subschema)

					for sub_field, json_path := range media_map {
						res := gjson.GetBytes(json_data, fmt.Sprintf(json_path, i, j))
						if incoming_value := res.Map(); len(incoming_value) > 0 {
							string_res := ""
							for key := range incoming_value {
								string_res += incoming_value[key].String() + " "
							}
							string_res = string_res[:len(string_res)-1]
							temp_subschema[sub_field] = string_res
						} else {
							temp_subschema[sub_field] = res.String()
						}
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
	return list_schema
}

func fillNormalTag(json_data []byte, schema map[string]any) map[string]any {
	for field, value := range schema {
		switch real_value := value.(type) {
		case string:
			res := gjson.GetBytes(json_data, real_value)
			schema[field] = res.String()
		}
	}
	return schema
}

func Translate(json_data []byte, requiredTags map[string]any, repeatTags map[string]string) map[string]any {

	resultData := addRepeatedTags(repeatTags)

	for tag, schema := range transformMap(requiredTags) {
		if list_json_path, ok := repeatTags[tag]; ok {
			list_schema, _ := resultData[tag].([]map[string]any)
			resultData[tag] = fillRepeatedTag(json_data, list_json_path, list_schema, schema)
		} else {
			resultData[tag] = fillNormalTag(json_data, schema)
		}
	}

	return resultData
}
