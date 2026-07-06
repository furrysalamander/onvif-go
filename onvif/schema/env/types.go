package env

import (
	"encoding/xml"
	"io"
	"strings"
)

const NS = "http://www.w3.org/2003/05/soap-envelope"

type Envelope struct {
	XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
	Header  *Header  `xml:"Header,omitempty"`
	Body    Body     `xml:"http://www.w3.org/2003/05/soap-envelope Body"`
}

type Header struct {
	XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Header"`
	Any     []rawXML `xml:",any"`
}

type Body struct {
	XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Body"`
	Fault   *Fault   `xml:"Fault,omitempty"`
	Any     []rawXML `xml:",any"`
}

type Fault struct {
	Code   *FaultCode   `xml:"Code"`
	Reason *FaultReason `xml:"Reason"`
	Node   string       `xml:"Node,omitempty"`
	Role   string       `xml:"Role,omitempty"`
	Detail *FaultDetail `xml:"Detail,omitempty"`
}

type FaultCode struct {
	Value   string     `xml:"Value"`
	Subcode *FaultCode `xml:"Subcode,omitempty"`
}

type FaultReason struct {
	Texts []FaultText `xml:"Text"`
}

type FaultText struct {
	Lang string `xml:"xml lang,attr"`
	Text string `xml:",chardata"`
}

type FaultDetail struct {
	Any []rawXML `xml:",any"`
}

type rawXML []byte

func (r rawXML) MarshalXML(e *xml.Encoder, se xml.StartElement) error {
	if len(r) == 0 {
		return nil
	}
	return pipeXML(e, r)
}

func (r *rawXML) UnmarshalXML(d *xml.Decoder, se xml.StartElement) error {
	var b strings.Builder
	enc := xml.NewEncoder(&b)
	if err := enc.EncodeToken(se); err != nil {
		return err
	}
	depth := 1
	for depth > 0 {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		if err := enc.EncodeToken(tok); err != nil {
			return err
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
	if err := enc.Flush(); err != nil {
		return err
	}
	*r = rawXML(b.String())
	return nil
}

func pipeXML(enc *xml.Encoder, r rawXML) error {
	if len(r) == 0 {
		return nil
	}
	dec := xml.NewDecoder(strings.NewReader(string(r)))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := enc.EncodeToken(tok); err != nil {
			return err
		}
	}
}

func (e *Envelope) SetBody(v interface{}) error {
	raw, err := xml.Marshal(v)
	if err != nil {
		return err
	}
	e.Body.Any = append(e.Body.Any, rawXML(raw))
	return nil
}

func (e *Envelope) AddHeader(v interface{}) error {
	if e.Header == nil {
		e.Header = &Header{}
	}
	raw, err := xml.Marshal(v)
	if err != nil {
		return err
	}
	e.Header.Any = append(e.Header.Any, rawXML(raw))
	return nil
}

func (b *Body) UnmarshalBody(v interface{}) error {
	if b.Fault != nil {
		return nil
	}
	if len(b.Any) == 0 {
		return nil
	}
	return xml.Unmarshal(b.Any[0], v)
}
