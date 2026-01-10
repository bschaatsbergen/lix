package view

import "fmt"

// ExportData contains the export success information to be rendered.
type ExportData struct {
	ImageRef   string
	OutputPath string
}

type ExportView interface {
	Render(data *ExportData) error
}

// Human view implementation
type exportHumanView struct {
	*HumanView
}

func newExportHumanView(hv *HumanView) *exportHumanView {
	return &exportHumanView{HumanView: hv}
}

func (v *exportHumanView) Render(data *ExportData) error {
	v.Printf("Exported %s to %s\n", data.ImageRef, data.OutputPath)
	return nil
}

// JSON view implementation
type exportJSONView struct {
	*JSONView
}

func newExportJSONView(jv *JSONView) *exportJSONView {
	return &exportJSONView{JSONView: jv}
}

func (v *exportJSONView) Render(data *ExportData) error {
	// TODO: implement JSON rendering when needed
	return fmt.Errorf("JSON view not yet implemented for export")
}
