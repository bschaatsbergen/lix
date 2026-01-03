package view

import (
	"encoding/json"
	"fmt"
)

// TagsData contains the list of tags to be rendered.
type TagsData struct {
	Repository string
	Tags       []string
}

type TagsView interface {
	Render(data *TagsData) error
}

// Human view implementation
type tagsHumanView struct {
	*HumanView
}

func newTagsHumanView(hv *HumanView) *tagsHumanView {
	return &tagsHumanView{HumanView: hv}
}

func (v *tagsHumanView) Render(data *TagsData) error {
	if len(data.Tags) == 0 {
		v.Printf("No tags found for %s\n", data.Repository)
		return nil
	}

	for _, tag := range data.Tags {
		v.Printf("%s\n", tag)
	}

	return nil
}

// JSON view implementation
type tagsJSONView struct {
	*JSONView
}

func newTagsJSONView(jv *JSONView) *tagsJSONView {
	return &tagsJSONView{JSONView: jv}
}

func (v *tagsJSONView) Render(data *TagsData) error {
	type jsonOutput struct {
		Repository string   `json:"repository"`
		Tags       []string `json:"tags,omitempty"`
		Message    string   `json:"message,omitempty"`
	}

	output := jsonOutput{
		Repository: data.Repository,
	}

	if len(data.Tags) == 0 {
		output.Message = fmt.Sprintf("No tags found for %s", data.Repository)
	} else {
		output.Tags = data.Tags
	}

	encoder := json.NewEncoder(v.Writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}
