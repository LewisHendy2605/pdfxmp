# PdfXmp - Flexible XMP Metadata Library for PDFs

GoXMP is a lightweight Go library to **read, extract, and manipulate XMP metadata** from PDFs. It is designed to be **flexible**, allowing users to handle **any XMP schema** and marshal it into their own types.

---

## Features

- Embed XMP metadata to existing PDFs
- Extract XMP metadata from any PDF
- Parse `rdf:Description` blocks individually
- Expose raw XML for user-defined schemas
- Supports custom namespaces
- Minimal dependencies

---

## Installation

```bash
go get github.com/LewisHendy2605/pdfxmp