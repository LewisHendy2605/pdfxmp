package templates

var byte_code = "\uFEFF"

var catalog = "1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n/Metadata %d 0 R\n>>\nendobj\n"

var xmp_obj = `%d 0 obj
<<
/Type /Metadata
/Subtype /XML
/Length %d
>>`
var xmp_stream = `
stream
%s
endstream
endobj%s`
var xmp_xml = "<?xpacket begin='\uFEFF' id='W5M0MpCehiHzreSzNTczkc9d'?>\n" +
	`<x:xmpmeta xmlns:x='adobe:ns:meta/' x:xmptk='Image::ExifTool 13.52'>
<rdf:RDF xmlns:rdf='http://www.w3.org/1999/02/22-rdf-syntax-ns#'>

 <rdf:Description rdf:about=''
  xmlns:dc='http://purl.org/dc/elements/1.1/'>
  <dc:creator>
   <rdf:Seq>
    <rdf:li>John Doe</rdf:li>
   </rdf:Seq>
  </dc:creator>
  <dc:title>
   <rdf:Alt>
    <rdf:li xml:lang='x-default'>Test PDF</rdf:li>
   </rdf:Alt>
  </dc:title>
</rdf:Description>

<rdf:Description rdf:about=''
 xmlns:pdf='http://ns.adobe.com/pdf/1.3/'>

 <pdf:Keywords>golang,xmp,pdf</pdf:Keywords>
 <pdf:Producer>My Go Tool</pdf:Producer>

</rdf:Description>

<rdf:Description rdf:about='' xmlns:my='http://example.com/custom/'>
	%s
</rdf:Description>

</rdf:RDF>
</x:xmpmeta>                                                                                                
<?xpacket end='w'?>`

var xref = `xref
0 2
0000000000 65535 f 
%010d 00000 n 
%d 1
%010d 00000 n 
trailer
<<
/Size %d
/Root 1 0 R
/Prev %d
>>
%s
startxref
%v
`
