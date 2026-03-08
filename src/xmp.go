package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/LewisHendy2605/pdfxmp/src/types"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// Load in string templates
var catalog, catalog_err = os.ReadFile("../templates/xmp_catalog.txt")
var xmp_obj, obj_err = os.ReadFile("../templates/xmp_obj.txt")
var xmp_stream, stream_err = os.ReadFile("../templates/xmp_stream.txt")
var xmp_xml, xml_err = os.ReadFile("../templates/xmp.xml")
var xref, xref_err = os.ReadFile("../templates/xref.txt")

type Pdf struct {
	inFilePath               string
	outFilePath              string
	xmpMetadata              string
	buffer                   bytes.Buffer
	numObjects               int
	prevStartXrefBytesOffset int
}

type Metadata struct {
	Key   string
	Value string
}

// Opens a existing pdf file and loads into memory
// TODO: Is this the best way to do this, whats the performace loss of holding in memory ?
func Open(filePath string) (*Pdf, error) {
	var err error

	if !strings.HasSuffix(filePath, ".pdf") {
		return nil, fmt.Errorf("file path had unexpected file extension")
	}

	path, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	pdf := &Pdf{
		inFilePath: path,
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = io.Copy(&pdf.buffer, file)
	if err != nil {
		return nil, err
	}

	pdf.scanObjects()
	err = pdf.prevStartXrefByteOffset()
	if err != nil {
		return nil, err
	}

	return pdf, nil
}

// Flushes buffer into a output file
func (pdf *Pdf) Save(filePath string) error {
	var err error

	path, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	pdf.outFilePath = path

	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, &pdf.buffer)
	if err != nil {
		return err
	}

	return nil
}

// Counts the ammout of objects in a pdf
func (pdf *Pdf) scanObjects() {
	re := regexp.MustCompile(`\d+\s+\d+\s+obj`)
	matches := re.FindAll(pdf.buffer.Bytes(), -1)
	pdf.numObjects = len(matches)
}

// Get the current start xref from the pdf
func (pdf *Pdf) prevStartXrefByteOffset() error {
	var err error

	re := regexp.MustCompile(`startxref\s+(\d+)`)
	matches := re.FindAllSubmatch(pdf.buffer.Bytes(), -1)

	if len(matches) == 0 {
		return fmt.Errorf("no prvious start xfref found")
	}

	last := matches[len(matches)-1]
	pdf.prevStartXrefBytesOffset, err = strconv.Atoi(string(last[1]))
	if err != nil {
		return err
	}

	return nil
}

func (pdf *Pdf) Validate() error {
	return api.ValidateFile(pdf.inFilePath, nil)
}

// Embeded xmp xml metadata into a pdf
func (pdf *Pdf) AddMetaData(metadata []Metadata) error {
	var err error

	if stream_err != nil {
		return err
	}

	// Add marker comment
	if _, err = pdf.buffer.Write([]byte("%BeginGpdfUpdate\n")); err != nil {
		return err
	}

	// Mark byte offset for catalog
	offsetCatalog := pdf.buffer.Len()

	// Write catalog
	if _, err = fmt.Fprintf(&pdf.buffer, string(catalog), pdf.numObjects+1); err != nil {
		return err
	}

	pdf.buffer.Write([]byte("\n"))

	// Build xml with embeded data
	xmp_xml_custom := buildXml(metadata)
	xml := fmt.Sprintf(string(xmp_xml), xmp_xml_custom)
	xmp_metadata := fmt.Sprintf(string(xmp_stream), xml)

	// Mark byte offset for xmp object
	offsetXMP := pdf.buffer.Len()

	// Write xmp object
	if _, err = fmt.Fprintf(&pdf.buffer, string(xmp_obj), pdf.numObjects+1, len(xmp_xml), xmp_metadata); err != nil {
		return err
	}

	pdf.buffer.Write([]byte("\n"))

	// Mark byte offset for xref table
	xrefOffset := pdf.buffer.Len()

	// Write xref table
	if _, err = fmt.Fprintf(&pdf.buffer, string(xref), offsetCatalog, pdf.numObjects+1, offsetXMP, pdf.numObjects+2, pdf.prevStartXrefBytesOffset, "%EndGpdfUpdate", xrefOffset); err != nil {
		return err
	}

	// Write end of file
	if _, err = pdf.buffer.Write([]byte("\n%%EOF\n")); err != nil {
		return err
	}

	return nil
}

// Builds xml with embeded metadata
func buildXml(metadata []Metadata) []byte {
	var xml strings.Builder
	for _, data := range metadata {
		xml.WriteString(newXmlTag(data.Key, data.Value) + "\n")
	}

	return []byte(xml.String())
}

// Creates a new xml tag
func newXmlTag(key string, value string) string {
	return fmt.Sprintf("<my:%s>%s</my:%s>", key, value, key)
}

// Extracts metadata from existing pdf and outputs as a file
func ExtractMetadata(inputFile string, outputDir string) error {
	path, err := filepath.Abs(outputDir)
	if err != nil {
		return err
	}
	err = api.ExtractMetadataFile(inputFile, path, nil)
	if err != nil {
		return fmt.Errorf("couldnt extract metadate from pdf: %w", err)
	}
	return nil
}

func GetMetadata(inputFile string) (*types.Metadata, error) {
	outputDir, err := filepath.Abs("../output")
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return nil, err
	}

	err = api.ExtractMetadataFile(inputFile, outputDir, nil)
	if err != nil {
		return nil, fmt.Errorf("couldnt extract metadate from pdf: %w", err)
	}

	pdfName := strings.Trim(filepath.Base(inputFile), filepath.Ext(inputFile))

	pattern := filepath.Join(
		outputDir,
		pdfName+"_Metadata_Catalog_*.txt",
	)

	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("No Metdata Files Found")
	}

	data, err := os.ReadFile(files[0])
	if err != nil {
		return nil, fmt.Errorf("No Metdata Files Found")
	}

	var meta types.Metadata

	err = xml.Unmarshal(data, &meta)
	if err != nil {
		return nil, err
	}

	return &meta, nil
}

func Validate(filePath string) error {
	return api.ValidateFile(filePath, nil)
}
