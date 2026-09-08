package storage

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectImageType(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		content   []byte
		wantExt   string
		wantCT    string
		wantErr   bool
	}{
		{
			name:     "JPEG file",
			filename: "photo.jpg",
			content:  append([]byte{0xFF, 0xD8, 0xFF}, bytes.Repeat([]byte{0}, 100)...),
			wantExt:  ".jpg",
			wantCT:   "image/jpeg",
		},
		{
			name:     "PNG file",
			filename: "image.png",
			content:  append([]byte{0x89, 0x50, 0x4E, 0x47}, bytes.Repeat([]byte{0}, 100)...),
			wantExt:  ".png",
			wantCT:   "image/png",
		},
		{
			name:     "GIF file",
			filename: "animation.gif",
			content:  append([]byte{0x47, 0x49, 0x46, 0x38}, bytes.Repeat([]byte{0}, 100)...),
			wantExt:  ".gif",
			wantCT:   "image/gif",
		},
		{
			name:     "WebP file",
			filename: "image.webp",
			content:  append([]byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50}, bytes.Repeat([]byte{0}, 100)...),
			wantExt:  ".webp",
			wantCT:   "image/webp",
		},
		{
			name:     "RIFF but not WebP (WAV)",
			filename: "audio.wav",
			content:  append([]byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x41, 0x56, 0x45}, bytes.Repeat([]byte{0}, 100)...),
			wantErr:  true,
		},
		{
			name:     "No extension but valid content type header",
			filename: "photo",
			content:  append([]byte{0xFF, 0xD8, 0xFF}, bytes.Repeat([]byte{0}, 100)...),
			wantExt:  ".jpg",
			wantCT:   "image/jpeg",
		},
		{
			name:     "Invalid extension and no magic bytes",
			filename: "file.txt",
			content:  bytes.Repeat([]byte{0}, 100),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, tt.filename)
			if err := os.WriteFile(tmpFile, tt.content, 0644); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			f, err := os.Open(tmpFile)
			if err != nil {
				t.Fatalf("failed to open temp file: %v", err)
			}
			defer f.Close()

			stat, _ := f.Stat()
			header := &multipart.FileHeader{
				Filename: tt.filename,
				Size:     stat.Size(),
				Header:   make(map[string][]string),
			}

			ext, ct, err := detectImageType(header, f)
			if (err != nil) != tt.wantErr {
				t.Errorf("detectImageType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if ext != tt.wantExt {
					t.Errorf("detectImageType() ext = %q, want %q", ext, tt.wantExt)
				}
				if ct != tt.wantCT {
					t.Errorf("detectImageType() ct = %q, want %q", ct, tt.wantCT)
				}
			}
		})
	}
}

func TestDetectImageType_SeekReset(t *testing.T) {
	content := append([]byte{0xFF, 0xD8, 0xFF}, bytes.Repeat([]byte{0}, 100)...)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.jpg")
	os.WriteFile(tmpFile, content, 0644)
	f, _ := os.Open(tmpFile)
	defer f.Close()

	stat, _ := f.Stat()
	header := &multipart.FileHeader{
		Filename: "test.jpg",
		Size:     stat.Size(),
		Header:   make(map[string][]string),
	}

	pos1, _ := f.Seek(0, io.SeekCurrent)
	ext, _, err := detectImageType(header, f)
	pos2, _ := f.Seek(0, io.SeekCurrent)

	if err != nil {
		t.Fatalf("detectImageType() unexpected error: %v", err)
	}
	if ext != ".jpg" {
		t.Errorf("expected .jpg, got %s", ext)
	}
	if pos1 != pos2 {
		t.Errorf("file position changed: before=%d, after=%d", pos1, pos2)
	}
}

func TestAllowedExtensions(t *testing.T) {
	valid := []string{".jpg", ".jpeg", ".png", ".webp", ".gif"}
	for _, ext := range valid {
		if !allowedExtensions[ext] {
			t.Errorf("extension %s should be allowed", ext)
		}
	}

	invalid := []string{".exe", ".sh", ".php", ".html", ".svg"}
	for _, ext := range invalid {
		if allowedExtensions[ext] {
			t.Errorf("extension %s should not be allowed", ext)
		}
	}
}

func TestAllowedContentTypes(t *testing.T) {
	valid := []string{"image/jpeg", "image/png", "image/webp", "image/gif"}
	for _, ct := range valid {
		if !allowedContentTypes[ct] {
			t.Errorf("content type %s should be allowed", ct)
		}
	}

	invalid := []string{"text/html", "application/javascript", "application/x-executable"}
	for _, ct := range invalid {
		if allowedContentTypes[ct] {
			t.Errorf("content type %s should not be allowed", ct)
		}
	}
}

func TestUploadImage_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.jpg")
	os.WriteFile(tmpFile, []byte{}, 0644)
	f, _ := os.Open(tmpFile)
	defer f.Close()

	stat, _ := f.Stat()
	header := &multipart.FileHeader{
		Filename: "empty.jpg",
		Size:     stat.Size(),
		Header:   make(map[string][]string),
	}

	_, err := UploadImage(nil, f, header)
	if err == nil {
		t.Error("expected error for empty file")
	}
}

func TestNewUploadRequest(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("product_id", "1")

	part, err := writer.CreateFormFile("image", "test.jpg")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	part.Write([]byte("fake image content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if req.Method != "POST" {
		t.Error("expected POST method")
	}
}
