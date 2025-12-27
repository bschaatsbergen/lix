package command

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/bschaatsbergen/lix/internal/oci"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/spf13/cobra"
)

type CompareOptions struct {
	Platform string
	Pull     string
}

func NewCompareCommand(cli *CLI) *cobra.Command {
	var opts CompareOptions

	cmd := &cobra.Command{
		Use:   "compare <image:tag1> <image:tag2>",
		Short: "Compare two OCI images with different tags",
		Long: `Compare two OCI images from the same repository with different tags.

Shows differences in layers, size, and configuration between the two images.

Examples:
  lix compare alpine:latest alpine:3.17
  lix compare nginx:1.25 nginx:1.24
  lix compare myregistry.io/app:v1.0 myregistry.io/app:v2.0`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunCompare(cmd.Context(), cli, args[0], args[1], &opts)
		},
	}

	cmd.Flags().StringVar(&opts.Platform, "platform", "", "Platform to fetch (e.g., linux/amd64, linux/arm64)")
	cmd.Flags().StringVar(&opts.Pull, "pull", "if-not-present", "Pull policy: always, if-not-present, never")

	return cmd
}

func RunCompare(ctx context.Context, cli *CLI, image1Ref, image2Ref string, opts *CompareOptions) error {
	logger := cli.Logger()

	// Parse both image references
	ref1, err := name.ParseReference(image1Ref)
	if err != nil {
		return fmt.Errorf("failed to parse first image reference: %w", err)
	}

	ref2, err := name.ParseReference(image2Ref)
	if err != nil {
		return fmt.Errorf("failed to parse second image reference: %w", err)
	}

	// Validate that both images are from the same repository
	repo1 := ref1.Context().Name()
	repo2 := ref2.Context().Name()

	if repo1 != repo2 {
		return fmt.Errorf("images must be from the same repository: %s != %s", repo1, repo2)
	}

	logger.Info("Comparing images", "image1", image1Ref, "image2", image2Ref)

	// Fetch both images
	fetchOpts := &oci.FetchOptions{
		Platform:   opts.Platform,
		PullPolicy: oci.PullPolicy(opts.Pull),
	}

	img1, _, err := oci.FetchImage(ctx, image1Ref, fetchOpts)
	if err != nil {
		return fmt.Errorf("failed to fetch first image: %w", err)
	}

	img2, _, err := oci.FetchImage(ctx, image2Ref, fetchOpts)
	if err != nil {
		return fmt.Errorf("failed to fetch second image: %w", err)
	}

	// Check if images are identical
	digest1, err := img1.Digest()
	if err != nil {
		return fmt.Errorf("failed to get digest for first image: %w", err)
	}

	digest2, err := img2.Digest()
	if err != nil {
		return fmt.Errorf("failed to get digest for second image: %w", err)
	}

	if digest1.String() == digest2.String() {
		cli.Printf("Images are identical\n")
		return nil
	}

	// Get layers from both images
	layers1, err := img1.Layers()
	if err != nil {
		return fmt.Errorf("failed to get layers from first image: %w", err)
	}

	layers2, err := img2.Layers()
	if err != nil {
		return fmt.Errorf("failed to get layers from second image: %w", err)
	}

	// Build layer digest sets to identify shared layers
	layerDigests1 := make(map[string]bool)
	layerDigests2 := make(map[string]bool)

	for _, layer := range layers1 {
		digest, err := layer.Digest()
		if err != nil {
			return fmt.Errorf("failed to get layer digest: %w", err)
		}
		layerDigests1[digest.String()] = true
	}

	for _, layer := range layers2 {
		digest, err := layer.Digest()
		if err != nil {
			return fmt.Errorf("failed to get layer digest: %w", err)
		}
		layerDigests2[digest.String()] = true
	}

	// Extract files only from unique layers (skip shared layers entirely)
	files1, err := extractFileListFromLayers(layers1, layerDigests2)
	if err != nil {
		return fmt.Errorf("failed to extract files from first image: %w", err)
	}

	files2, err := extractFileListFromLayers(layers2, layerDigests1)
	if err != nil {
		return fmt.Errorf("failed to extract files from second image: %w", err)
	}

	// Compare file lists
	added, removed, modified := compareFiles(files1, files2)

	// Display results
	if len(added) > 0 {
		cli.Printf("Added:\n")
		for _, path := range added {
			cli.Printf("  %s\n", path)
		}
	}

	if len(removed) > 0 {
		cli.Printf("Removed:\n")
		for _, path := range removed {
			cli.Printf("  %s\n", path)
		}
	}

	if len(modified) > 0 {
		cli.Printf("Modified:\n")
		for _, path := range modified {
			cli.Printf("  %s\n", path)
		}
	}

	if len(added) == 0 && len(removed) == 0 && len(modified) == 0 {
		cli.Printf("No file changes detected\n")
	}

	return nil
}

// extractFileListFromLayers builds the filesystem state, skipping shared layers
func extractFileListFromLayers(layers []v1.Layer, otherLayerDigests map[string]bool) (map[string]bool, error) {
	files := make(map[string]bool)

	// Process layers from bottom to top (overlay filesystem)
	for _, layer := range layers {
		layerDigest, err := layer.Digest()
		if err != nil {
			return nil, fmt.Errorf("failed to get layer digest: %w", err)
		}

		digestStr := layerDigest.String()

		// Skip shared layers
		if otherLayerDigests[digestStr] {
			continue
		}

		// This layer is unique, read its files
		rc, err := layer.Uncompressed()
		if err != nil {
			return nil, fmt.Errorf("failed to get uncompressed layer: %w", err)
		}

		tr := tar.NewReader(rc)
		for {
			header, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				rc.Close()
				return nil, fmt.Errorf("failed to read tar header: %w", err)
			}

			// Normalize path
			path := "/" + strings.TrimPrefix(header.Name, "/")

			// Skip directories
			if header.Typeflag == tar.TypeDir {
				continue
			}

			// Handle whiteout files (deletions)
			if strings.HasPrefix(header.Name, ".wh.") {
				delete(files, path)
				continue
			}

			// Only process regular files
			if header.Typeflag == tar.TypeReg {
				files[path] = true
				// Skip file content
				if _, err := io.Copy(io.Discard, tr); err != nil {
					rc.Close()
					return nil, fmt.Errorf("failed to skip file contents: %w", err)
				}
			}
		}
		rc.Close()
	}

	return files, nil
}

func compareFiles(files1, files2 map[string]bool) (added, removed, modified []string) {
	// Since we only extracted files from unique layers: - Files in both maps
	// are from different unique layers = modified - Files only in files2 =
	// added - Files only in files1 = removed

	for path := range files2 {
		if _, exists := files1[path]; exists {
			// File in both unique layer sets = modified
			modified = append(modified, path)
		} else {
			// File only in image2's unique layers = added
			added = append(added, path)
		}
	}

	// Find removed files
	for path := range files1 {
		if _, exists := files2[path]; !exists {
			removed = append(removed, path)
		}
	}

	// Sort for consistent output
	sort.Strings(added)
	sort.Strings(removed)
	sort.Strings(modified)

	return
}
