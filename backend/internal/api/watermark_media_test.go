package api

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
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
	if err := generateWatermarkedImage(source, target, "DALU PARTS | TEST VENDOR", 32, "image/png"); err != nil {
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
}

func TestWatermarkLabelTransliteratesChineseAndPathStaysInMediaRoot(t *testing.T) {
	label := watermarkLabel("大陆农机配件", "汉丰")
	if !strings.Contains(label, "da") || !strings.Contains(label, "han") {
		t.Fatalf("watermark label was not transliterated: %q", label)
	}
	root := t.TempDir()
	if _, err := safeMediaPath(root, "..\\outside.png"); err == nil {
		t.Fatal("unsafe media path was accepted")
	}
}
