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


func mapClone[M ~map[K]V, K comparable, V any](dict M) M {
	new_dict := M{}
	for key, value := range dict {
		new_dict[key] = value
	}
	return new_dict
}

func translate(json_data []byte, requiredTags map[string]any, repeatTags map[string]string) map[string]any {
	
	resultData := map[string]any{}
	for tag := range repeatTags {
		resultData[tag] = []map[string]string{}
	}

	for tag, schema := range transformMap(requiredTags){
		if list_json_path, ok := repeatTags[tag]; ok {
			list_schema, _ := resultData[tag].([]map[string]any)
			for i := range gjson.GetBytes(json_data, list_json_path).Array() {
				temp_schema := mapClone(schema)
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
							temp_subschema := mapClone(subschema)
							
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