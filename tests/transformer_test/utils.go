package transformer

func assertMaps(actual map[string]any, expected map[string]any) bool {
	if len(actual) != len(expected) {
		return false
	}
	for key, raw_value := range actual {
		if value, ok := actual[key].(map[string]any); ok {
			if expected_value, ok := expected[key].(map[string]any); ok {
				if !assertMaps(value, expected_value){
					return false
				}
			} else {
				return false
			}
		} else if value, ok := actual[key].([]map[string]any); ok{
			if expected_value, ok := expected[key].([]map[string]any); ok {
				if len(value) != len(expected_value) {
					return false
				}
				for i := range value {
					if !assertMaps(value[i], expected_value[i]) {
						return false
					}
				}
			} else {
				return false
			}
		}else {
			if raw_value != expected[key] {
				return false
			}
		}
	}
	return  true
}


var ComplexTemplate string = `<#Repeat#Story#>
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
			</#Repeat#Story#>`

var ComplexExpectedTemplate string = `{{ range $Story := .Story }}
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
			{{ end }}`
