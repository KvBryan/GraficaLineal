package document_analysis

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestParseWordFile(t *testing.T) {
	// Generate in-memory docx zip with word/document.xml
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	xmlContent := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>Introducción y Metodología de la Investigación</w:t></w:r></w:p>
    <w:tbl>
      <w:tr>
        <w:tc><w:p><w:r><w:t>Grupo</w:t></w:r></w:p></w:tc>
        <w:tc><w:p><w:r><w:t>Puntuación</w:t></w:r></w:p></w:tc>
      </w:tr>
      <w:tr>
        <w:tc><w:p><w:r><w:t>Control</w:t></w:r></w:p></w:tc>
        <w:tc><w:p><w:r><w:t>65.4</w:t></w:r></w:p></w:tc>
      </w:tr>
      <w:tr>
        <w:tc><w:p><w:r><w:t>Experimental</w:t></w:r></w:p></w:tc>
        <w:tc><w:p><w:r><w:t>92.1</w:t></w:r></w:p></w:tc>
      </w:tr>
    </w:tbl>
  </w:body>
</w:document>`

	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatalf("Failed to create xml in zip: %v", err)
	}
	w.Write([]byte(xmlContent))
	zw.Close()

	res, err := ParseWordFile(buf.Bytes(), "test_document.docx")
	if err != nil {
		t.Fatalf("ParseWordFile error: %v", err)
	}

	if res.FileType != "word" {
		t.Errorf("Expected fileType word, got %s", res.FileType)
	}
	if len(res.Tables) != 1 {
		t.Fatalf("Expected 1 table, got %d", len(res.Tables))
	}
	if len(res.Tables[0].Headers) != 2 {
		t.Errorf("Expected 2 headers, got %d", len(res.Tables[0].Headers))
	}
	if res.Tables[0].Headers[0] != "Grupo" || res.Tables[0].Headers[1] != "Puntuación" {
		t.Errorf("Unexpected headers: %v", res.Tables[0].Headers)
	}
	if len(res.Tables[0].PreviewRows) != 2 {
		t.Errorf("Expected 2 preview rows, got %d", len(res.Tables[0].PreviewRows))
	}
}
