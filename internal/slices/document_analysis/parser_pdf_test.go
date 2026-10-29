package document_analysis

import (
	"os"
	"os/exec"
	"testing"
)

func TestParsePDFFile(t *testing.T) {
	// Generate valid sample PDF using ps2pdf if not existing
	pdfPath := "/tmp/sample_unit_test.pdf"
	psContent := `%!PS
/Helvetica findfont 12 scalefont setfont
72 700 moveto
(Tratamiento   Dosis   Eficacia) show
72 680 moveto
(Control       0mg     24.5) show
72 660 moveto
(Activo        50mg    89.2) show
showpage`

	psPath := "/tmp/sample_unit_test.ps"
	if err := os.WriteFile(psPath, []byte(psContent), 0644); err == nil {
		cmd := exec.Command("ps2pdf", psPath, pdfPath)
		_ = cmd.Run()
	}

	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Skip("ps2pdf not available, skipping live PDF generation test")
		return
	}

	res, err := ParsePDFFile(pdfBytes, "ensayo_clinico.pdf")
	if err != nil {
		t.Fatalf("ParsePDFFile failed: %v", err)
	}

	if res.FileType != "pdf" {
		t.Errorf("Expected fileType pdf, got %s", res.FileType)
	}

	if len(res.Sections) == 0 {
		t.Errorf("Expected at least 1 section in PDF")
	}
}
