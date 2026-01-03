package view

import (
	"encoding/json"
	"fmt"
)

// CatData contains the file content to be rendered.
type CatData struct {
	Content string
}

type CatView interface {
	Render(data *CatData) error
}

// Human view implementation
type catHumanView struct {
	*HumanView
}

func newCatHumanView(hv *HumanView) *catHumanView {
	return &catHumanView{HumanView: hv}
}

func (v *catHumanView) Render(data *CatData) error {
	v.Printf("%s", data.Content)
	return nil
}

// JSON view implementation
type catJSONView struct {
	*JSONView
}

func newCatJSONView(jv *JSONView) *catJSONView {
	return &catJSONView{JSONView: jv}
}

func (v *catJSONView) Render(data *CatData) error {
	type jsonOutput struct {
		Content string `json:"content"`
	}

	output := jsonOutput{
		Content: data.Content,
	}

	encoder := json.NewEncoder(v.Writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}
