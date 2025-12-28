package view

import (
	"fmt"

	"github.com/bschaatsbergen/cek/internal/oci"
)

// CompareData contains the comparison information to be rendered.
type CompareData struct {
	Image1Ref    string
	Image2Ref    string
	Image1Layers int
	Image2Layers int
	Image1Size   int64
	Image2Size   int64
	Added        []string
	Removed      []string
	Modified     []string
	Identical    bool
}

type CompareView interface {
	Render(data *CompareData) error
}

// Human view implementation
type compareHumanView struct {
	*HumanView
}

func newCompareHumanView(hv *HumanView) *compareHumanView {
	return &compareHumanView{HumanView: hv}
}

func (v *compareHumanView) Render(data *CompareData) error {
	if data.Identical {
		v.Printf("Images are identical\n")
		return nil
	}

	v.Printf("Image 1: %s\n", data.Image1Ref)
	v.Printf("  Layers: %d\n", data.Image1Layers)
	v.Printf("  Size: %s\n\n", oci.FormatBytes(data.Image1Size))

	v.Printf("Image 2: %s\n", data.Image2Ref)
	v.Printf("  Layers: %d\n", data.Image2Layers)
	v.Printf("  Size: %s\n\n", oci.FormatBytes(data.Image2Size))

	if len(data.Added) > 0 {
		v.Printf("Added:\n")
		for _, path := range data.Added {
			v.Printf("  %s\n", path)
		}
		v.Printf("\n")
	}

	if len(data.Removed) > 0 {
		v.Printf("Removed:\n")
		for _, path := range data.Removed {
			v.Printf("  %s\n", path)
		}
		v.Printf("\n")
	}

	if len(data.Modified) > 0 {
		v.Printf("Modified:\n")
		for _, path := range data.Modified {
			v.Printf("  %s\n", path)
		}
		v.Printf("\n")
	}

	if len(data.Added) == 0 && len(data.Removed) == 0 && len(data.Modified) == 0 {
		v.Printf("No file changes detected\n")
	}

	return nil
}

// JSON view implementation
type compareJSONView struct {
	*JSONView
}

func newCompareJSONView(jv *JSONView) *compareJSONView {
	return &compareJSONView{JSONView: jv}
}

func (v *compareJSONView) Render(data *CompareData) error {
	// TODO: implement JSON rendering when needed
	return fmt.Errorf("JSON view not yet implemented for compare")
}
