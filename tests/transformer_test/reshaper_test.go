package transformer

import (
	"PrintServer/internal/transformer"
	"testing"
)

func TestChangeRepeatTags(t *testing.T) {
	test_data := []struct {
		TestTemplate       string
		ExpectedTemplate   string
		ExpectedRepeats    []string
		ExpectedRepeatTags map[string]string
	}{
		{
			"<#Repeat#Story#></#Repeat#Story#>",
			"{{ range $Story := .Story }}{{ end }}",
			[]string{"Story"},
			map[string]string{
				"Story": "rundown.content",
			},
		},
		{
			"<#Repeat#Story#><#Repeat#Media#></#Repeat#Media#></#Repeat#Story#>",
			"{{ range $Story := .Story }}{{ range $Media := .Media }}{{ end }}{{ end }}",
			[]string{"Story", "Media"},
			map[string]string{
				"Story": "rundown.content",
				"Media": "rundown.content.%d.story.media_content",
			},
		},
		{
			"<#Repeat#Story#><#Repeat#Media#><#Repeat#Unknown#></#Repeat#Unknown#></#Repeat#Media#></#Repeat#Story#>",
			"{{ range $Story := .Story }}{{ range $Media := .Media }}{{ range $Unknown := .Unknown }}{{ end }}{{ end }}{{ end }}",
			[]string{"Story", "Media", "Unknown"},
			map[string]string{
				"Story":   "rundown.content",
				"Media":   "rundown.content.%d.story.media_content",
				"Unknown": "",
			},
		},
	}

	for _, set := range test_data {
		repeatTags := map[string]string{}
		repeats, actual_template_content := test_reshaper.ChangeRepeatTags(
			test_context,
			set.TestTemplate,
			&repeatTags,
		)
		if len(repeats) != len(set.ExpectedRepeats) {
			t.Errorf("Wrong repeats expected: %v got: %v", set.ExpectedRepeats, repeats)
		}
		if actual_template_content != set.ExpectedTemplate {
			t.Errorf("Wrong reshape expected:\n %v\n got:\n %v\n", set.ExpectedTemplate, actual_template_content)
		}
		if len(repeatTags) != len(set.ExpectedRepeatTags) {
			t.Errorf("Wrong repeat tags expected: %v got: %v", set.ExpectedRepeatTags, repeatTags)
		}
	}

}

func TestIncludeRepeatStructure(t *testing.T) {
	test_data := []struct {
		TestTemplate         string
		ExpectedRequiredTags map[string]any
		ExpectedConnections  map[string]any
	}{
		{
			"<#Repeat#Story#></#Repeat#Story#>",
			map[string]any{
				"Story": map[string]any{},
			},
			map[string]any{},
		},
		{
			"<#Repeat#Story#><#Repeat#Media#></#Repeat#Media#></#Repeat#Story#>",
			map[string]any{
				"Story": map[string]any{
					"Media": map[string]any{},
				},
			},
			map[string]any{
				"Story": []string{"Media"},
			},
		},
		{
			"<#Repeat#Story#></#Repeat#Story#><#Repeat#Media#></#Repeat#Media#>",
			map[string]any{
				"Story": map[string]any{},
				"Media": map[string]any{},
			},
			map[string]any{},
		},
		{
			"<#Repeat#Story#><#Repeat#Media#><#Repeat#Unknown#></#Repeat#Unknown#></#Repeat#Media#></#Repeat#Story#>",
			map[string]any{
				"Story": map[string]any{
					"Media": map[string]any{
						"Unknown": map[string]any{},
					},
				},
			},
			map[string]any{
				"Story": []string{"Media"},
				"Media": []string{"Unknown"},
			},
		},
		{
			"<#Repeat#Story#><#Repeat#Media#></#Repeat#Media#></#Repeat#Story#><#Repeat#Unknown#></#Repeat#Unknown#>",
			map[string]any{
				"Story": map[string]any{
					"Media": map[string]any{},
				},
				"Unknown": map[string]any{},
			},
			map[string]any{
				"Story": []string{"Media"},
			},
		},
	}

	for _, set := range test_data {
		requiredTags := map[string]any{}
		connections := transformer.IncludeRepeatStructure(
			set.TestTemplate,
			&requiredTags,
		)
		checkable_connections := make(map[string]any, len(connections))
		for src, dest := range connections {
			checkable_connections[src] = dest
		}
		if !assertMaps(checkable_connections, set.ExpectedConnections) {
			t.Errorf("Wrong connections expected: %v got: %v", set.ExpectedConnections, connections)
		}
		if !assertMaps(requiredTags, set.ExpectedRequiredTags) {
			t.Errorf("Wrong requiredTags expected:\n %v\n got:\n %v\n", set.ExpectedRequiredTags, requiredTags)
		}
	}
}

func TestChangeBaseTags(t *testing.T) {
	test_data := []struct {
		TestTemplate         string
		ExpectedTemplate     string
		ExpectedRequiredTags map[string]any
	}{
		{
			ComplexTemplate,
			ComplexExpectedTemplate,
			map[string]any{
				"Story": map[string]any{
					"Pos":          "rundown.content.%d.pos",
					"Name":         "rundown.content.%d.story.name",
					"Type":         "rundown.content.%d.story.type",
					"Author":       "rundown.content.%d.story.author",
					"State":        "rundown.content.%d.story.status",
					"Presenter":    "rundown.content.%d.story.presenter",
					"DurationPlan": "rundown.content.%d.story.plan_durat",
					"StartPlan":    "rundown.content.%d.story.plan_start",
					"DurationFact": "rundown.content.%d.story.fact_duration",
					"Text":         "rundown.content.%d.story.text.text",
					"Media":        map[string]any{},
				},
				"Media": map[string]any{
					"Pos":    "rundown.content.%d.story.media_content.%d.pos",
					"Type":   "rundown.content.%d.story.media_content.%d.type",
					"Name":   "rundown.content.%d.story.media_content.%d.name",
					"Start":  "rundown.content.%d.story.media_content.%d.mark_in",
					"Params": "rundown.content.%d.story.media_content.%d",
				},
			},
		},
		{
			`<#Repeat#Story#>
			#Story.Pos#
			#Story.Name#
			#Story.Type#
			</#Repeat#Story#>
			<#Repeat#Media#>
			#Media.Pos#
			</#Repeat#Media#>
			`,
			`{{ range $Story := .Story }}
			{{ $Story.Pos }}
			{{ $Story.Name }}
			{{ $Story.Type }}
			{{ end }}
			{{ range $Media := .Media }}
			{{ $Media.Pos }}
			{{ end }}
			`,
			map[string]any{
				"Story": map[string]any{
					"Pos":  "rundown.content.%d.pos",
					"Name": "rundown.content.%d.story.name",
					"Type": "rundown.content.%d.story.type",
				},
				"Media": map[string]any{
					"Pos": "rundown.content.%d.story.media_content.%d.pos",
				},
			},
		},
		{
			`#RunDown.Author#
			#Story.Pos#
			#Story.Name#
			#Story.Type#
			#Media.Pos#
			`,
			`{{ .RunDown.Author }}
			{{ .Story.Pos }}
			{{ .Story.Name }}
			{{ .Story.Type }}
			{{ .Media.Pos }}
			`,
			map[string]any{
				"RunDown": map[string]any{
					"Author": "rundown.author",
				},
				"Story": map[string]any{
					"Pos":  "rundown.content.%d.pos",
					"Name": "rundown.content.%d.story.name",
					"Type": "rundown.content.%d.story.type",
				},
				"Media": map[string]any{
					"Pos": "rundown.content.%d.story.media_content.%d.pos",
				},
			},
		},
		{
			`#RunDown.Author.Unknown#
			`,
			`#RunDown.Author.Unknown#
			`,
			map[string]any{},
		},
	}
	for _, set := range test_data {
		requiredTags := map[string]any{}
		repeatTags := map[string]string{}
		transformer.IncludeRepeatStructure(set.TestTemplate, &requiredTags)
		repeats, template_content := test_reshaper.ChangeRepeatTags(test_context, set.TestTemplate, &repeatTags)
		template_content = test_reshaper.ChangeBaseTags(test_context, template_content, repeats, &requiredTags)
		if template_content != set.ExpectedTemplate {
			t.Errorf("Wrong reshape expected:\n %v \n got:\n %v \n", set.ExpectedTemplate, template_content)
		}

		if !assertMaps(requiredTags, set.ExpectedRequiredTags) {
			t.Errorf("Wrong requiredTags expected:\n %v \n got:\n %v \n", set.ExpectedRequiredTags, requiredTags)
		}
	}
}

func TestCompleteConnections(t *testing.T) {
	test_data := []struct {
		TestTemplate         string
		ExpectedRequiredTags map[string]any
		ExpectedRepeatTags   map[string]any
	}{
		{
			ComplexTemplate,
			map[string]any{
				"Story": map[string]any{
					"Pos":          "rundown.content.%d.pos",
					"Name":         "rundown.content.%d.story.name",
					"Type":         "rundown.content.%d.story.type",
					"Author":       "rundown.content.%d.story.author",
					"State":        "rundown.content.%d.story.status",
					"Presenter":    "rundown.content.%d.story.presenter",
					"DurationPlan": "rundown.content.%d.story.plan_durat",
					"StartPlan":    "rundown.content.%d.story.plan_start",
					"DurationFact": "rundown.content.%d.story.fact_duration",
					"Text":         "rundown.content.%d.story.text.text",
					"Media": map[string]any{
						"Pos":    "rundown.content.%d.story.media_content.%d.pos",
						"Type":   "rundown.content.%d.story.media_content.%d.type",
						"Name":   "rundown.content.%d.story.media_content.%d.name",
						"Start":  "rundown.content.%d.story.media_content.%d.mark_in",
						"Params": "rundown.content.%d.story.media_content.%d",
					},
				},
			},
			map[string]any{
				"Story": "rundown.content",
			},
		},
		{
			`#RunDown.Author#
			#Story.Pos#
			#Story.Name#
			#Story.Type#
			#Media.Pos#
			`,
			map[string]any{
				"RunDown": map[string]any{
					"Author": "rundown.author",
				},
				"Story": map[string]any{
					"Pos":  "rundown.content.%d.pos",
					"Name": "rundown.content.%d.story.name",
					"Type": "rundown.content.%d.story.type",
				},
				"Media": map[string]any{
					"Pos": "rundown.content.%d.story.media_content.%d.pos",
				},
			},
			map[string]any{},
		},
		{
			`<#Repeat#Story#>
			#Story.Pos#
			#Story.Name#
			#Story.Type#
			</#Repeat#Story#>
			<#Repeat#Media#>
			#Media.Pos#
			</#Repeat#Media#>
			`,
			map[string]any{
				"Story": map[string]any{
					"Pos":  "rundown.content.%d.pos",
					"Name": "rundown.content.%d.story.name",
					"Type": "rundown.content.%d.story.type",
				},
				"Media": map[string]any{
					"Pos": "rundown.content.%d.story.media_content.%d.pos",
				},
			},
			map[string]any{
				"Story": "rundown.content",
				"Media": "rundown.content.%d.story.media_content",
			},
		},
	}

	for _, set := range test_data {
		requiredTags := map[string]any{}
		repeatTags := map[string]string{}
		connections := transformer.IncludeRepeatStructure(set.TestTemplate, &requiredTags)
		repeats, template_content := test_reshaper.ChangeRepeatTags(test_context, set.TestTemplate, &repeatTags)
		test_reshaper.ChangeBaseTags(test_context, template_content, repeats, &requiredTags)
		transformer.CompleteConnections(connections, &requiredTags, &repeatTags)

		if !assertMaps(requiredTags, set.ExpectedRequiredTags) {
			t.Errorf("Wrong requiredTags expected:\n %v \n got:\n %v \n", set.ExpectedRequiredTags, requiredTags)
		}

		checkable_repeatTags := make(map[string]any, len(repeatTags))
		for src, dest := range repeatTags {
			checkable_repeatTags[src] = dest
		}

		if !assertMaps(checkable_repeatTags, set.ExpectedRepeatTags) {
			t.Errorf("Wrong requiredTags expected:\n %v \n got:\n %v \n", set.ExpectedRepeatTags, checkable_repeatTags)
		}
	}

}
