package pdfxmp

import (
	"bytes"
	"embed"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	types "github.com/LewisHendy2605/pdfxmp/internal/extract"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

//go:embed templates/*
var templateFS embed.FS

// Load in string templates
var catalog, catalog_err = templateFS.ReadFile("templates/xmp_catalog.txt")
var xmp_obj, obj_err = templateFS.ReadFile("templates/xmp_obj.txt")
var xmp_stream, stream_err = templateFS.ReadFile("templates/xmp_stream.txt")
var xmp_xml, xml_err = templateFS.ReadFile("templates/xmp.xml")
var xref, xref_err = templateFS.ReadFile("templates/xref.txt")

var creator_xml, creator_err = templateFS.ReadFile("templates/creator.xml")
var title_xml, title_err = templateFS.ReadFile("templates/title.xml")

type Pdf struct {
	inFilePath           string
	outFilePath          string
	xmpMetadata          string
	buffer               bytes.Buffer
	pdfMetadata          bytes.Buffer
	adobeMetadata        bytes.Buffer
	numObjects           int
	startXrefBytesOffset int
}

type Metadata struct {
	Key   string
	Value string
}

// Opens a existing pdf file and loads into memory
// TODO: Whats the performace loss of holding in memory ?
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

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = io.Copy(&pdf.buffer, file)
	if err != nil {
		return nil, err
	}

	pdf.numObjects = getNumObj(pdf.buffer.Bytes())
	pdf.startXrefBytesOffset, err = getStartXrefByteOffset(pdf.buffer.Bytes())
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

// Validates pdf file with pdfcpu
func (pdf *Pdf) Validate() error {
	return api.ValidateFile(pdf.inFilePath, nil)
}

// Adds Description to xmp metatdata
func (pdf *Pdf) AddCreator(creator string) error {
	_, err := fmt.Fprintf(&pdf.pdfMetadata, string(creator_xml)+"\n", creator)
	if err != nil {
		return fmt.Errorf("couldnt embed creator matadata: %w", err)
	}
	return nil
}

// Adds Description to xmp metatdata
func (pdf *Pdf) AddTitle(title string) error {
	_, err := fmt.Fprintf(&pdf.pdfMetadata, string(title_xml)+"\n", title)
	if err != nil {
		return fmt.Errorf("couldnt embed title matadata: %w", err)
	}
	return nil
}

// Adds Keywords to xmp metatdata, keywords can be a comma seperate string
func (pdf *Pdf) AddKeyword(keywords string) error {
	_, err := pdf.adobeMetadata.Write([]byte(newXmlTag("Keywords", keywords, "pdf") + "\n"))
	if err != nil {
		return fmt.Errorf("couldnt embed keywords: %w", err)
	}
	return nil
}

// Adds Producer to xmp metatdata
func (pdf *Pdf) AddProducer(producer string) error {
	_, err := pdf.adobeMetadata.Write([]byte(newXmlTag("Producer", producer, "pdf") + "\n"))
	if err != nil {
		return fmt.Errorf("couldnt embed creator matadata: %w", err)
	}
	return nil
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
	xml := fmt.Sprintf(string(xmp_xml), pdf.pdfMetadata.String(), pdf.adobeMetadata.String(), xmp_xml_custom)
	xmp_metadata := fmt.Sprintf(string(xmp_stream), xml)

	// Mark byte offset for xmp object
	offsetXMP := pdf.buffer.Len()

	// Write xmp object
	if _, err = fmt.Fprintf(&pdf.buffer, string(xmp_obj), pdf.numObjects+1, len(xml), xmp_metadata); err != nil {
		return err
	}

	pdf.buffer.Write([]byte("\n"))

	// Mark byte offset for xref table
	xrefOffset := pdf.buffer.Len()

	// Write xref table
	if _, err = fmt.Fprintf(&pdf.buffer, string(xref), offsetCatalog, pdf.numObjects+1, offsetXMP, pdf.numObjects+2, pdf.startXrefBytesOffset, "%EndGpdfUpdate", xrefOffset); err != nil {
		return err
	}

	// Write end of file
	if _, err = pdf.buffer.Write([]byte("\n%%EOF\n")); err != nil {
		return err
	}

	return nil
}

// Counts the ammout of objects in a pdf
func getNumObj(bytes []byte) int {
	re := regexp.MustCompile(`\d+\s+\d+\s+obj`)
	matches := re.FindAll(bytes, -1)
	return len(matches)
}

// Get the current start xref from the pdf
func getStartXrefByteOffset(bytes []byte) (int, error) {
	var err error

	re := regexp.MustCompile(`startxref\s+(\d+)`)
	matches := re.FindAllSubmatch(bytes, -1)

	if len(matches) == 0 {
		return -1, fmt.Errorf("no prvious start xfref found")
	}

	last := matches[len(matches)-1]
	result, err := strconv.Atoi(string(last[1]))
	if err != nil {
		return -1, err
	}

	return result, nil
}

// Builds xml with embeded metadata
func buildXml(metadata []Metadata) []byte {
	var xml strings.Builder
	for _, data := range metadata {
		xml.WriteString(newXmlTag(data.Key, data.Value, "meta") + "\n")
	}

	return []byte(xml.String())
}

// Creates a new xml tag
// E.g <prefix:key>value</prefix:key>
func newXmlTag(key string, value string, prefix string) string {
	return fmt.Sprintf("<%s:%s>%s</%s:%s>", prefix, key, value, prefix, key)
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

// Extract metadata form a pdf, marshes xml data and returns the object as a struct
// Exposes the descriptiosn as a list with the raw xml to unmarshal to cutom types
func GetMetadata(inputFile string) (*types.Metadata, error) {
	outputDir, err := filepath.Abs("./output")
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

// Validates a pdf
func Validate(filePath string) error {
	return api.ValidateFile(filePath, nil)
}
