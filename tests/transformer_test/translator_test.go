package transformer

import (
	"former/internal/transformer"
	"os"
	"testing"
)

func TestTranslate(t *testing.T) {
	test_json, _ := os.ReadFile("test.json")
	test_data := []struct {
		RequiredTags map[string]any
		RepeatTags   map[string]string
		ExpectedData map[string]any
	}{
		{
			map[string]any{
				"Story": map[string]any{
					"Pos": "rundown.content.%d.pos",
					"Name": "rundown.content.%d.story.name",
					"Media": map[string]any{
						"Pos": "rundown.content.%d.story.media_content.%d.pos",
					},
				},
			},
			map[string]string{
				"Story": "rundown.content",
			},
			map[string]any{
				"Story": []map[string]any{
					{
						"Pos": "1",
						"Name": "ШПИГЕЛЬ",
						"Media": []map[string]any{
							{
								"Pos": "0",
							},
							{
								"Pos": "1",
							},
						},
					},
				},
			},
		},
		{
			map[string]any{
				"RunDown": map[string]any{
					"Name": "rundown.title",
					"ID": "rundown_id",
				},
				"Story": map[string]any{
					"Pos": "rundown.content.%d.pos",
					"Name": "rundown.content.%d.story.name",
				},
			},
			map[string]string{
				"Story": "rundown.content",
			},
			map[string]any{
				"RunDown": map[string]any{
					"Name": "НОВОСТИ 12:00",
					"ID": "2",
				},
				"Story": []map[string]any{
					{
						"Pos": "1",
						"Name": "ШПИГЕЛЬ",
					},
				},
			},
		},
	}
	for _, set := range test_data {
		data := transformer.Translate(test_json, set.RequiredTags, set.RepeatTags)
		if !assertMaps(data, set.ExpectedData) {
			t.Errorf("Wrong requiredTags expected:\n %v \n got:\n %v \n", set.ExpectedData, data)
		}
	}
}
