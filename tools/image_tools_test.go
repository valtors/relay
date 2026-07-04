package tools

import (
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageInfoReturnsDimensionsAndFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.png")
	writeTestImage(t, path)

	res, err := ImageInfo(t.Context(), makeImageReq(map[string]any{"path": path}))
	require.NoError(t, err)
	require.False(t, res.IsError)

	var payload struct {
		Width         int    `json:"width"`
		Height        int    `json:"height"`
		Format        string `json:"format"`
		FileSizeBytes int64  `json:"file_size_bytes"`
	}
	decodeImageResult(t, res, &payload)

	assert.Equal(t, 4, payload.Width)
	assert.Equal(t, 3, payload.Height)
	assert.Equal(t, "png", payload.Format)
	assert.Positive(t, payload.FileSizeBytes)
}

func TestImageResizeScalesProportionally(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.png")
	writeTestImage(t, path)

	res, err := ImageResize(t.Context(), makeImageReq(map[string]any{
		"path":  path,
		"width": 8,
	}))
	require.NoError(t, err)
	require.False(t, res.IsError)

	var payload struct {
		OutputPath string `json:"output_path"`
		Width      int    `json:"width"`
		Height     int    `json:"height"`
	}
	decodeImageResult(t, res, &payload)
	assert.Equal(t, 8, payload.Width)
	assert.Equal(t, 6, payload.Height)

	img, format, err := loadImage(payload.OutputPath)
	require.NoError(t, err)
	assert.Equal(t, "png", format)
	assert.Equal(t, 8, img.Bounds().Dx())
	assert.Equal(t, 6, img.Bounds().Dy())
}

func TestImageCropReturnsCroppedRegion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.png")
	writeTestImage(t, path)

	res, err := ImageCrop(t.Context(), makeImageReq(map[string]any{
		"path":   path,
		"x":      1,
		"y":      1,
		"width":  2,
		"height": 2,
	}))
	require.NoError(t, err)
	require.False(t, res.IsError)

	var payload struct {
		OutputPath string `json:"output_path"`
		Width      int    `json:"width"`
		Height     int    `json:"height"`
	}
	decodeImageResult(t, res, &payload)
	assert.Equal(t, 2, payload.Width)
	assert.Equal(t, 2, payload.Height)

	img, _, err := loadImage(payload.OutputPath)
	require.NoError(t, err)
	assertColorEquals(t, testColorAt(1, 1), img.At(0, 0))
	assertColorEquals(t, testColorAt(2, 2), img.At(1, 1))
}

func TestImageConvertRoundTripFormats(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.png")
	writeTestImage(t, path)

	jpegRes, err := ImageConvert(t.Context(), makeImageReq(map[string]any{
		"path":    path,
		"format":  "jpeg",
		"quality": 90,
	}))
	require.NoError(t, err)
	require.False(t, jpegRes.IsError)

	var jpegPayload struct {
		OutputPath string `json:"output_path"`
	}
	decodeImageResult(t, jpegRes, &jpegPayload)
	_, jpegFormat, err := loadImage(jpegPayload.OutputPath)
	require.NoError(t, err)
	assert.Equal(t, "jpeg", jpegFormat)

	gifRes, err := ImageConvert(t.Context(), makeImageReq(map[string]any{
		"path":   jpegPayload.OutputPath,
		"format": "gif",
	}))
	require.NoError(t, err)
	require.False(t, gifRes.IsError)

	var gifPayload struct {
		OutputPath string `json:"output_path"`
	}
	decodeImageResult(t, gifRes, &gifPayload)
	_, gifFormat, err := loadImage(gifPayload.OutputPath)
	require.NoError(t, err)
	assert.Equal(t, "gif", gifFormat)
}

func TestImageRotateVariants(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.png")
	writeTestImage(t, path)

	cases := []struct {
		name        string
		degrees     int
		wantW       int
		wantH       int
		topLeft     color.Color
		bottomRight color.Color
	}{
		{
			name:        "rotate90",
			degrees:     90,
			wantW:       3,
			wantH:       4,
			topLeft:     testColorAt(0, 2),
			bottomRight: testColorAt(3, 0),
		},
		{
			name:        "rotate180",
			degrees:     180,
			wantW:       4,
			wantH:       3,
			topLeft:     testColorAt(3, 2),
			bottomRight: testColorAt(0, 0),
		},
		{
			name:        "rotate270",
			degrees:     270,
			wantW:       3,
			wantH:       4,
			topLeft:     testColorAt(3, 0),
			bottomRight: testColorAt(0, 2),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := ImageRotate(t.Context(), makeImageReq(map[string]any{
				"path":    path,
				"degrees": tc.degrees,
			}))
			require.NoError(t, err)
			require.False(t, res.IsError)

			var payload struct {
				OutputPath string `json:"output_path"`
			}
			decodeImageResult(t, res, &payload)

			img, _, err := loadImage(payload.OutputPath)
			require.NoError(t, err)
			assert.Equal(t, tc.wantW, img.Bounds().Dx())
			assert.Equal(t, tc.wantH, img.Bounds().Dy())
			assertColorEquals(t, tc.topLeft, img.At(0, 0))
			assertColorEquals(t, tc.bottomRight, img.At(img.Bounds().Dx()-1, img.Bounds().Dy()-1))
		})
	}
}

func TestImageGrayscaleProducesGrayPixels(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.png")
	writeTestImage(t, path)

	res, err := ImageGrayscale(t.Context(), makeImageReq(map[string]any{"path": path}))
	require.NoError(t, err)
	require.False(t, res.IsError)

	var payload struct {
		OutputPath string `json:"output_path"`
	}
	decodeImageResult(t, res, &payload)

	img, _, err := loadImage(payload.OutputPath)
	require.NoError(t, err)

	r, g, b, _ := img.At(1, 1).RGBA()
	assert.Equal(t, r, g)
	assert.Equal(t, g, b)
}

func TestImageFlipDirections(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.png")
	writeTestImage(t, path)

	hRes, err := ImageFlip(t.Context(), makeImageReq(map[string]any{
		"path":      path,
		"direction": "horizontal",
	}))
	require.NoError(t, err)
	require.False(t, hRes.IsError)

	var hPayload struct {
		OutputPath string `json:"output_path"`
	}
	decodeImageResult(t, hRes, &hPayload)
	hImg, _, err := loadImage(hPayload.OutputPath)
	require.NoError(t, err)
	assertColorEquals(t, testColorAt(3, 0), hImg.At(0, 0))
	assertColorEquals(t, testColorAt(0, 2), hImg.At(3, 2))

	vRes, err := ImageFlip(t.Context(), makeImageReq(map[string]any{
		"path":      path,
		"direction": "vertical",
	}))
	require.NoError(t, err)
	require.False(t, vRes.IsError)

	var vPayload struct {
		OutputPath string `json:"output_path"`
	}
	decodeImageResult(t, vRes, &vPayload)
	vImg, _, err := loadImage(vPayload.OutputPath)
	require.NoError(t, err)
	assertColorEquals(t, testColorAt(0, 2), vImg.At(0, 0))
	assertColorEquals(t, testColorAt(3, 0), vImg.At(3, 2))
}

func TestImageErrors(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		res, err := ImageInfo(t.Context(), makeImageReq(map[string]any{"path": filepath.Join(t.TempDir(), "missing.png")}))
		require.NoError(t, err)
		require.True(t, res.IsError)
		assert.Contains(t, textOf(t, res), "load image")
	})

	t.Run("invalid format", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "sample.png")
		writeTestImage(t, path)

		res, err := ImageConvert(t.Context(), makeImageReq(map[string]any{
			"path":   path,
			"format": "bmp",
		}))
		require.NoError(t, err)
		require.True(t, res.IsError)
		assert.Contains(t, textOf(t, res), "format must be png, jpeg, or gif")
	})

	t.Run("crop outside bounds", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "sample.png")
		writeTestImage(t, path)

		res, err := ImageCrop(t.Context(), makeImageReq(map[string]any{
			"path":   path,
			"x":      3,
			"y":      2,
			"width":  2,
			"height": 2,
		}))
		require.NoError(t, err)
		require.True(t, res.IsError)
		assert.Contains(t, textOf(t, res), "outside image bounds")
	})
}

func makeImageReq(args map[string]any) mcp.CallToolRequest {
	var req mcp.CallToolRequest
	req.Params.Name = "image_tool"
	req.Params.Arguments = args
	return req
}

func decodeImageResult(t *testing.T, res *mcp.CallToolResult, target any) {
	t.Helper()
	require.NoError(t, json.Unmarshal([]byte(textOf(t, res)), target))
}

func writeTestImage(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 3))
	for y := 0; y < 3; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, testColorAt(x, y))
		}
	}

	file, err := os.Create(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = file.Close() })
	require.NoError(t, png.Encode(file, img))
	require.NoError(t, file.Close())
}

func testColorAt(x, y int) color.Color {
	return color.RGBA{
		R: uint8(20 + x*50),
		G: uint8(30 + y*70),
		B: uint8(40 + (x+y)*20),
		A: 255,
	}
}

func assertColorEquals(t *testing.T, want, got color.Color) {
	t.Helper()
	assert.Equal(t, color.NRGBAModel.Convert(want), color.NRGBAModel.Convert(got))
}
