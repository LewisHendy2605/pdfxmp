package pdfxmp

import (
	"fmt"
	"testing"
	"time"
)

func TestXmpEmbeding(t *testing.T) {
	var err error
	inputPdf := "../examples/blank_one_page.pdf"
	outputPdf := fmt.Sprintf("../output/%s.pdf", time.Now().Format("02-01-2006"))

	pdf, err := Open(inputPdf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = pdf.Validate()

	// Build MetaData
	var metadata []Metadata
	metadata = append(metadata, Metadata{Key: "Plex", Value: "Simplex"})
	metadata = append(metadata, Metadata{Key: "PostageClass", Value: "Economy"})

	// Embed
	err = pdf.AddMetaData(metadata)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = pdf.Save(outputPdf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = Validate(outputPdf)
	if err != nil {
		t.Fatalf("unepected error: %v", err)
	}

	_, err = GetMetadata(outputPdf)
	if err != nil {
		t.Fatalf("unepected error: %v", err)
	}
}
