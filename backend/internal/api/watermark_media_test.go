package api

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateWatermarkedImageCreatesSeparateDerivative(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source.png")
	target := filepath.Join(root, "public-watermarks", "display.png")
	file, err := os.Create(source)
	if err != nil {
		t.Fatal(err)
	}
	base := image.NewRGBA(image.Rect(0, 0, 240, 120))
	for y := 0; y < 120; y++ {
		for x := 0; x < 240; x++ {
			base.Set(x, y, color.RGBA{R: 220, G: 230, B: 240, A: 255})
		}
	}
	if err := png.Encode(file, base); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := generateWatermarkedImage(source, target, "大陆农机配件 · 汉丰", 25, "image/png"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("watermarked derivative not created: %v", err)
	}
	sourceBytes, _ := os.ReadFile(source)
	targetBytes, _ := os.ReadFile(target)
	if string(sourceBytes) == string(targetBytes) {
		t.Fatal("derivative is identical to source")
	}
	output, err := os.Open(target)
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	watermarked, _, err := image.Decode(output)
	if err != nil {
		t.Fatal(err)
	}
	minChannel, maxChannel := uint32(255), uint32(0)
	bounds := watermarked.Bounds()
	for y := maxInt(bounds.Min.Y, bounds.Max.Y-42); y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, _, _, _ := watermarked.At(x, y).RGBA()
			channel := r >> 8
			if channel < minChannel {
				minChannel = channel
			}
			if channel > maxChannel {
				maxChannel = channel
			}
		}
	}
	if minChannel >= 120 || maxChannel <= 235 {
		t.Fatalf("watermark contrast is too weak: min=%d max=%d", minChannel, maxChannel)
	}
}

func TestWatermarkLabelKeepsChineseAndPathStaysInMediaRoot(t *testing.T) {
	label := watermarkLabel("大陆农机配件", "汉丰")
	if label != "大陆农机配件 · 汉丰" {
		t.Fatalf("unexpected watermark label: %q", label)
	}
	root := t.TempDir()
	for _, storageKey := range []string{"../outside.png", "..\\outside.png", "/outside.png", "C:\\outside.png"} {
		if _, err := safeMediaPath(root, storageKey); err == nil {
			t.Fatalf("unsafe media path was accepted: %q", storageKey)
		}
	}
	valid, err := safeMediaPath(root, "public-watermarks\\display.png")
	if err != nil {
		t.Fatalf("valid nested media path was rejected: %v", err)
	}
	if valid != filepath.Join(root, "public-watermarks", "display.png") {
		t.Fatalf("unexpected normalized media path: %q", valid)
	}
}
