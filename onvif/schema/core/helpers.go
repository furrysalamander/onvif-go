package core

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type QName struct {
	NS    string
	Local string
}

func (q QName) String() string {
	if q.NS == "" {
		return q.Local
	}
	return "{" + q.NS + "}" + q.Local
}

func (q QName) MarshalXML(e *xml.Encoder, se xml.StartElement) error {
	return e.EncodeElement(q.String(), se)
}

func (q *QName) UnmarshalXML(d *xml.Decoder, se xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &se); err != nil {
		return err
	}
	*q = parseQNameCanonical(s)
	return nil
}

func parseQNameCanonical(s string) QName {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "{") {
		if i := strings.Index(s, "}"); i > 0 {
			return QName{NS: s[1:i], Local: s[i+1:]}
		}
	}
	return QName{Local: s}
}

type Duration struct {
	Negative bool
	Years    int
	Months   int
	Days     int
	Hours    int
	Minutes  int
	Seconds  float64
}

func (d Duration) String() string {
	var b strings.Builder
	if d.Negative {
		b.WriteByte('-')
	}
	b.WriteByte('P')
	if d.Years != 0 {
		fmt.Fprintf(&b, "%dY", d.Years)
	}
	if d.Months != 0 {
		fmt.Fprintf(&b, "%dM", d.Months)
	}
	if d.Days != 0 {
		fmt.Fprintf(&b, "%dD", d.Days)
	}
	if d.Hours != 0 || d.Minutes != 0 || d.Seconds != 0 {
		b.WriteByte('T')
		if d.Hours != 0 {
			fmt.Fprintf(&b, "%dH", d.Hours)
		}
		if d.Minutes != 0 {
			fmt.Fprintf(&b, "%dM", d.Minutes)
		}
		if d.Seconds != 0 {
			fmt.Fprintf(&b, "%gS", d.Seconds)
		}
	}
	return b.String()
}

func (d Duration) MarshalXML(e *xml.Encoder, se xml.StartElement) error {
	return e.EncodeElement(d.String(), se)
}

func (d *Duration) UnmarshalXML(dec *xml.Decoder, se xml.StartElement) error {
	var s string
	if err := dec.DecodeElement(&s, &se); err != nil {
		return err
	}
	parsed, err := ParseDuration(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

func ParseDuration(s string) (Duration, error) {
	d := Duration{}
	if s == "" {
		return d, nil
	}
	if s[0] == '-' {
		d.Negative = true
		s = s[1:]
	}
	if !strings.HasPrefix(s, "P") {
		return d, fmt.Errorf("xsd:duration: missing leading P in %q", s)
	}
	s = s[1:]
	timeIdx := strings.Index(s, "T")
	datePart := s
	timePart := ""
	if timeIdx >= 0 {
		datePart, timePart = s[:timeIdx], s[timeIdx+1:]
	}
	if err := scanInts(datePart, map[string]*int{"Y": &d.Years, "M": &d.Months, "D": &d.Days}); err != nil {
		return d, err
	}
	rem := timePart
	for len(rem) > 0 {
		n := 0
		for n < len(rem) && (rem[n] == '-' || rem[n] == '+' || (rem[n] >= '0' && rem[n] <= '9') || rem[n] == '.') {
			n++
		}
		if n == 0 || n == len(rem) {
			return d, fmt.Errorf("xsd:duration: malformed time part %q", timePart)
		}
		num := rem[:n]
		unit := rem[n]
		switch unit {
		case 'H':
			v, err := strconv.Atoi(num)
			if err != nil {
				return d, fmt.Errorf("xsd:duration: bad hours %q", num)
			}
			d.Hours = v
		case 'M':
			v, err := strconv.Atoi(num)
			if err != nil {
				return d, fmt.Errorf("xsd:duration: bad minutes %q", num)
			}
			d.Minutes = v
		case 'S':
			v, err := strconv.ParseFloat(num, 64)
			if err != nil {
				return d, fmt.Errorf("xsd:duration: bad seconds %q", num)
			}
			d.Seconds = v
		default:
			return d, fmt.Errorf("xsd:duration: unknown time unit %q", unit)
		}
		rem = rem[n+1:]
	}
	return d, nil
}

func scanInts(part string, dsts map[string]*int) error {
	rem := part
	for len(rem) > 0 {
		best := -1
		var bestUnit string
		for unit := range dsts {
			i := strings.Index(rem, unit)
			if i > 0 && (best == -1 || i < best) {
				best = i
				bestUnit = unit
			}
		}
		if best == -1 {
			return nil
		}
		n, err := strconv.Atoi(rem[:best])
		if err != nil {
			return fmt.Errorf("xsd:duration: bad integer %q", rem[:best])
		}
		*dsts[bestUnit] = n
		rem = rem[best+len(bestUnit):]
	}
	return nil
}

type RawXML []byte

func pipeXML(enc *xml.Encoder, r RawXML) error {
	if len(r) == 0 {
		return nil
	}
	dec := xml.NewDecoder(strings.NewReader(string(r)))
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
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

type Extension struct {
	AnyAttributes []xml.Attr
	Any           []RawXML
}

func (x Extension) MarshalXML(e *xml.Encoder, se xml.StartElement) error {
	_ = se
	if len(x.Any) == 0 && len(x.AnyAttributes) == 0 {
		return nil
	}
	for _, frag := range x.Any {
		if err := pipeXML(e, frag); err != nil {
			return err
		}
	}
	return nil
}

func (x *Extension) UnmarshalXML(d *xml.Decoder, se xml.StartElement) error {
	for _, a := range se.Attr {
		if a.Name.Space == "xmlns" {
			continue
		}
		x.AnyAttributes = append(x.AnyAttributes, a)
	}
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		if _, ok := tok.(xml.EndElement); ok {
			return nil
		}
		if start, ok := tok.(xml.StartElement); ok {
			frag, err := bufferSubtree(d, start)
			if err != nil {
				return err
			}
			x.Any = append(x.Any, frag)
		}
	}
}

func bufferSubtree(d *xml.Decoder, start xml.StartElement) (RawXML, error) {
	var b strings.Builder
	enc := xml.NewEncoder(&b)
	if err := enc.EncodeToken(start); err != nil {
		return nil, err
	}
	depth := 1
	for depth > 0 {
		tok, err := d.Token()
		if err != nil {
			return nil, err
		}
		if err := enc.EncodeToken(tok); err != nil {
			return nil, err
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
	if err := enc.Flush(); err != nil {
		return nil, err
	}
	return RawXML(b.String()), nil
}
