package two

import (
	"encoding/json"
	"fmt"
	"os"
)

type Rundown struct {
	ID       int         `json:"rundown_id"`
	StudioId int         `json:"studio_id"`
	Rundown  RundownData `json:"rundown"`
	Command  string      `json:"command"`
	Id       int         `json:"id"`
}

type RundownData struct {
	Content  []ContentElement `json:"content"`
	Title    string           `json:"title"`
	AirDate  string           `json:"air_date"`
	Duration int              `json:"duration"`
}

type ContentElement struct {
	ObjectId  int               `json:"obj_id"`
	TypeName  string            `json:"type_name"`
	Name      string            `json:"name"`
	Position  int               `json:"pos"`
	SkipFlag  int               `json:"skip_flag"`
	StartType int               `json:"start_type"`
	Note      string            `json:"note"`
	Story     *StoryModel       `json:"story"`
	SndPrompt int               `json:"snd_prompt"`
	Content   *[]ContentElement `json:"content"`
}

type StoryModel struct {
	Id                       int           `json:"id"`
	Name                     string        `json:"name"`
	Type                     string        `json:"type"`
	Version                  int           `json:"version"`
	Presenter                int           `json:"presenter"`
	PlanDuration             int           `json:"plan_durat"`
	TextDuration             int           `json:"text_duration"`
	CommentsDuration         int           `json:"comments_duration"`
	CommentsPrompterDuration int           `json:"comments_prompter_duration"`
	GapDuration              int           `json:"gap_duration"`
	VideoDuration            int           `json:"video_duration"`
	GraphicsDuration         int           `json:"graphics_duration"`
	FactDuration             int           `json:"fact_duration"`
	Status                   string        `json:"status"`
	StatusVideo              string        `json:"status_video"`
	MediaContent             *[]MediaModel `json:"media_content"`
	Text                     TextModel     `json:"text"`
}

type TextModel struct {
	TextXML string        `json:"text_xml"`
	Text    string        `json:"text"`
	Blocks  *[]BlockModel `json:"blocks"`
}

type BlockModel struct {
	BlockId     int            `json:"block_id"`
	ContentData []BlockContent `json:"content_data"`
	Position    int            `json:"pos"`
}

type BlockContent struct {
	Text        string `json:"text"`
	Duration    int    `json:"duration"`
	ContentType int    `json:"content_type"`
}

type MediaModel struct {
	AssetId  int    `json:"asset_id"`
	Name     string `json:"name"`
	Type     int    `json:"type"`
	Duration int    `json:"duration"`
	MarkIn   int    `json:"mark_in"`
	MarkOut  int    `json:"mark_out"`
}

func main() {
	data, err := os.ReadFile("rundown_print_2.json")
	if err != nil {
		fmt.Println("Pizdec", err.Error())
	}
	var some_data Rundown
	err = json.Unmarshal(data, &some_data)
	if err != nil {
		fmt.Println("Pizdec", err.Error())
	}
	fmt.Println(some_data.Rundown.Title)
	for i, element := range some_data.Rundown.Content {
		if element.TypeName == "story" {
			fmt.Println("START OF ELEMENT")
			fmt.Println("Pos:", i)
			fmt.Println("DurationFact:", element.Story.FactDuration)
			fmt.Println("Type:", element.Story.Type)
			fmt.Println("Name:", element.Story.Name)
			fmt.Println("END OF ELEMENT")
		}
	}

}
