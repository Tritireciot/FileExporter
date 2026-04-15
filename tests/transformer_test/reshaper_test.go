package transformer

import (
	"former/internal/transformer"
	"testing"
)

func TestChangeRepeatTags(t *testing.T) {
	test_data := []struct {
		TestTemplate string
		ExpectedTemplate string
		ExpectedRepeats []string
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
				"Story": "rundown.content",
				"Media": "rundown.content.%d.story.media_content",
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
		TestTemplate string
		ExpectedRequiredTags map[string]any
		ExpectedConnections map[string]string
	}{
		{
			"<#Repeat#Story#></#Repeat#Story#>",
			map[string]any{
				"Story": map[string]any{},
			},
			map[string]string{},
		},
		{
			"<#Repeat#Story#><#Repeat#Media#></#Repeat#Media#></#Repeat#Story#>",
			map[string]any{
				"Story": map[string]any{
					"Media": map[string]any{},
				},
			},
			map[string]string{
				"Story": "Media",
			},
		},
		{
			"<#Repeat#Story#></#Repeat#Story#><#Repeat#Media#></#Repeat#Media#>",
			map[string]any{
				"Story": map[string]any{},
				"Media": map[string]any{},
			},
			map[string]string{},
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
			map[string]string{
				"Story": "Media",
				"Media": "Unknown",
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
			map[string]string{
				"Story": "Media",
			},
		},
	}

	for _, set := range test_data {
		requiredTags := map[string]any{}
		connections := transformer.IncludeRepeatStructure(
			set.TestTemplate, 
			&requiredTags,
		)
		if len(connections) != len(set.ExpectedConnections) {
			t.Errorf("Wrong connections expected: %v got: %v", set.ExpectedConnections, connections)
		}
		if len(requiredTags) != len(set.ExpectedRequiredTags) {
			t.Errorf("Wrong requiredTags expected:\n %v\n got:\n %v\n", set.ExpectedRequiredTags, requiredTags)
		}
	}
}

func TestChangeBaseTags(t *testing.T){
	test_data := []struct {
		TestTemplate string
		ExpectedTemplate string
		ExpectedRequiredTags map[string]any
	}{
		{
			`<#Repeat#Story#>
			<hr />
			<p style="text-align: left;">№#Story.Pos#. #Story.Name#</p>
			<table style="width: 100%; margin-left: auto; margin-right: auto;">
			<tbody>
			<tr>
			<td colspan="5" nowrap>#Story.Type#</td>
			</tr>
			<tr>
			<td>Автор:</td>
			<td nowrap>#Story.Author#</td>
			<td></td>
			<td>Готовность:</td>
			<td nowrap>#Story.State#</td>
			</tr>
			<tr>
			<td>Начитка:</td>
			<td nowrap>#Story.Presenter#</td>
			<td>&nbsp;</td>
			<td>&nbsp;</td>
			<td>&nbsp;</td>
			</tr>
			<tr>
			<td colspan="2" nowrap>Время старта от начала выпуска:</td>
			<td>&nbsp;</td>
			<td nowrap>Длительность планируемая:</td>
			<td nowrap>#Story.DurationPlan#</td>
			</tr>
			<tr>
			<td colspan="2" nowrap>#Story.StartPlan#</td>
			<td>&nbsp;</td>
			<td nowrap>Длительность фактическая:</td>
			<td nowrap>#Story.DurationFact#</td>
			</tr>
			</tbody>
			</table>
			<p style="word-wrap:break-word;">#Story.Text#</p>
			<table width="100%">
			<tbody>
			<#Repeat#Media#>
			<tr>
			<td nowrap>#Story.Pos# #Media.Pos#</td>
			<td nowrap>#Media.Type# #Media.Name#</td>
			<td align="right" nowrap>#Media.Start#</td>
			</tr>		
			<tr>
			<td colspan="2">#Media.Params#</td>
			</tr>
			</#Repeat#Media#>
			</tbody>
			</table>
			</#Repeat#Story#>`,
			`{{ range $Story := .Story }}
			<hr />
			<p style="text-align: left;">№{{ $Story.Pos }}. {{ $Story.Name }}</p>
			<table style="width: 100%; margin-left: auto; margin-right: auto;">
			<tbody>
			<tr>
			<td colspan="5" nowrap>{{ $Story.Type }}</td>
			</tr>
			<tr>
			<td>Автор:</td>
			<td nowrap>{{ $Story.Author }}</td>
			<td></td>
			<td>Готовность:</td>
			<td nowrap>{{ $Story.State }}</td>
			</tr>
			<tr>
			<td>Начитка:</td>
			<td nowrap>{{ $Story.Presenter }}</td>
			<td>&nbsp;</td>
			<td>&nbsp;</td>
			<td>&nbsp;</td>
			</tr>
			<tr>
			<td colspan="2" nowrap>Время старта от начала выпуска:</td>
			<td>&nbsp;</td>
			<td nowrap>Длительность планируемая:</td>
			<td nowrap>{{ $Story.DurationPlan }}</td>
			</tr>
			<tr>
			<td colspan="2" nowrap>{{ $Story.StartPlan }}</td>
			<td>&nbsp;</td>
			<td nowrap>Длительность фактическая:</td>
			<td nowrap>{{ $Story.DurationFact }}</td>
			</tr>
			</tbody>
			</table>
			<p style="word-wrap:break-word;">{{ $Story.Text }}</p>
			<table width="100%">
			<tbody>
			{{ range $Media := .Media }}
			<tr>
			<td nowrap>{{ $Story.Pos }} {{ $Media.Pos }}</td>
			<td nowrap>{{ $Media.Type }} {{ $Media.Name }}</td>
			<td align="right" nowrap>{{ $Media.Start }}</td>
			</tr>		
			<tr>
			<td colspan="2">{{ $Media.Params }}</td>
			</tr>
			{{ end }}
			</tbody>
			</table>
			{{ end }}`,
			map[string]any{
				"Story": map[string]any{
					"Pos": "rundown.content.%d.pos",
					"Name": "rundown.content.%d.story.id",
					"Type": "rundown.content.%d.story.type",
					"Author": "rundown.content.%d.story.author",
					"State": "rundown.content.%d.story.status",
					"Presenter": "rundown.content.%d.story.presenter",
					"DurationPlan": "rundown.content.%d.story.plan_durat",
					"StartPlan": "rundown.content.%d.story.plan_start",
					"DurationFact": "rundown.content.%d.story.fact_duration",
					"Text": "rundown.content.%d.story.text.text",
					"Media": map[string]any{},
				},
				"Media": map[string]any{
					"Pos": "rundown.content.%d.story.media_content.%d.pos",
					"Type": "rundown.content.%d.story.media_content.%d.type",
					"Name": "rundown.content.%d.story.media_content.%d.name",
					"Start": "rundown.content.%d.story.media_content.%d.mark_in",
					"Params": "rundown.content.%d.story.media_content.%d",
				},
			},
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

		if len(requiredTags) != len(set.ExpectedRequiredTags) {
			t.Errorf("Wrong requiredTags expected:\n %v \n got:\n %v \n", set.ExpectedRequiredTags, requiredTags)
		}
	}
}