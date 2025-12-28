package view

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bschaatsbergen/cek/internal/oci"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// InspectData contains all the information needed to rendered.
type InspectData struct {
	ImageRef     string
	Registry     string
	Digest       v1.Hash
	Created      time.Time
	OS           string
	Architecture string
	TotalSize    int64
	Layers       []LayerData
}

// LayerData contains information about a single layer.
type LayerData struct {
	Index  int
	Digest v1.Hash
	Size   int64
}

type InspectView interface {
	Render(data *InspectData) error
}

// Human view implementation
type inspectHumanView struct {
	*HumanView
}

func newInspectHumanView(hv *HumanView) *inspectHumanView {
	return &inspectHumanView{HumanView: hv}
}

func (v *inspectHumanView) Render(data *InspectData) error {
	v.Printf("Image: %s\n", data.ImageRef)
	v.Printf("Registry: %s\n", data.Registry)
	v.Printf("Digest: %s\n", data.Digest)
	v.Printf("Created: %s\n", data.Created.Format(time.RFC3339))
	v.Printf("OS/Arch: %s/%s\n", data.OS, data.Architecture)
	v.Printf("Size: %s\n", oci.FormatBytes(data.TotalSize))
	v.Printf("\n")
	v.Printf("Layers:\n")
	v.Printf("#   %-66s %s\n", "Digest", "Size")

	for _, layer := range data.Layers {
		v.Printf("%-3d %-66s %s\n", layer.Index, layer.Digest.String(), oci.FormatBytes(layer.Size))
	}

	return nil
}

// JSON view implementation
type inspectJSONView struct {
	*JSONView
}

func newInspectJSONView(jv *JSONView) *inspectJSONView {
	return &inspectJSONView{JSONView: jv}
}

func (v *inspectJSONView) Render(data *InspectData) error {
	type jsonLayer struct {
		Index  int    `json:"index"`
		Digest string `json:"digest"`
		Size   int64  `json:"size"`
	}

	type jsonOutput struct {
		Image    string      `json:"image"`
		Registry string      `json:"registry"`
		Digest   string      `json:"digest"`
		Created  string      `json:"created"`
		OS       string      `json:"os"`
		Arch     string      `json:"arch"`
		Size     int64       `json:"size"`
		Layers   []jsonLayer `json:"layers"`
	}

	layers := make([]jsonLayer, len(data.Layers))
	for i, layer := range data.Layers {
		layers[i] = jsonLayer{
			Index:  layer.Index,
			Digest: layer.Digest.String(),
			Size:   layer.Size,
		}
	}

	output := jsonOutput{
		Image:    data.ImageRef,
		Registry: data.Registry,
		Digest:   data.Digest.String(),
		Created:  data.Created.Format(time.RFC3339),
		OS:       data.OS,
		Arch:     data.Architecture,
		Size:     data.TotalSize,
		Layers:   layers,
	}

	encoder := json.NewEncoder(v.Writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}
