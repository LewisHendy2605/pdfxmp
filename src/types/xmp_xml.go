package types

import "encoding/xml"

type Metadata struct {
	XMLName xml.Name `xml:"xmpmeta"`
	RDF     RDF      `xml:"RDF"`
}

type RDF struct {
	Descriptions []RawDescription `xml:"Description"`
}

type RawDescription struct {
	XMLName  xml.Name
	InnerXML []byte `xml:",innerxml"`
}

type Creator struct {
	Seq Seq `xml:"Seq"`
}

type Title struct {
	Alt Alt `xml:"Alt"`
}

type Seq struct {
	Li []string `xml:"li"`
}

type Alt struct {
	Li []LangValue `xml:"li"`
}

type LangValue struct {
	Lang  string `xml:"lang,attr"`
	Value string `xml:",chardata"`
}
