package main

import (
	"strings"
	"testing"
)

// TestRenderPreviewSanitizesTextControlBytes verifies text previews cannot emit control bytes.
func TestRenderPreviewSanitizesTextControlBytes(t *testing.T) {
	preview, binary := renderPreview([]byte("hello\x1b[2Jworld"))
	if binary {
		t.Fatal("text preview marked as binary")
	}
	if strings.Contains(preview, "\x1b") {
		t.Fatalf("preview contains raw escape byte: %q", preview)
	}
	if !strings.Contains(preview, "hello.[2Jworld") {
		t.Fatalf("unexpected preview: %q", preview)
	}
}

// TestRenderPreviewHexDumpsBinary verifies binary previews are rendered as hex.
func TestRenderPreviewHexDumpsBinary(t *testing.T) {
	preview, binary := renderPreview([]byte{0x00, 0x01, 0xff, 0x10})
	if !binary {
		t.Fatal("binary preview not marked as binary")
	}
	if !strings.Contains(preview, "00000000") || strings.Contains(preview, "\x00") {
		t.Fatalf("unexpected binary preview: %q", preview)
	}
}

// TestFormatBytes verifies compact IEC byte formatting at unit boundaries.
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{name: "bytes", size: 512, want: "512 B"},
		{name: "one kibibyte", size: 1024, want: "1.0 KiB"},
		{name: "fractional kibibytes", size: 1536, want: "1.5 KiB"},
		{name: "mebibytes", size: 5 * 1024 * 1024, want: "5.0 MiB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatBytes(tt.size); got != tt.want {
				t.Fatalf("formatBytes(%d) = %q, want %q", tt.size, got, tt.want)
			}
		})
	}
}

// TestFlattenMetadata verifies metadata headers flatten to first display values.
func TestFlattenMetadata(t *testing.T) {
	got := flattenMetadata(map[string][]string{
		"x-amz-meta-owner": {"ops", "ignored"},
		"empty":            nil,
	})

	if got["x-amz-meta-owner"] != "ops" {
		t.Fatalf("owner metadata = %q, want ops", got["x-amz-meta-owner"])
	}
	if got["empty"] != "" {
		t.Fatalf("empty metadata = %q, want empty string", got["empty"])
	}
}
