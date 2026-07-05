// Package ir holds the intermediate representation produced by the ONVIF
// WSDL/XSD parser and consumed by the Go code generator.
//
// The IR is intentionally lossy about XML specifics (namespace prefixes,
// attribute order) but preserves everything the generator needs to emit Go:
// per-namespace types, elements, attributes, operations, messages, and the
// SOAP bindings that wire them to HTTP.
package ir

import "sort"

// QName is a (namespace, local-name) reference. Empty namespace means
// "unqualified / unknown / same schema".
type QName struct {
	NS    string
	Local string
}

// String returns the canonical "URI:Local" form used in golden dumps.
func (q QName) String() string {
	if q.NS == "" {
		return q.Local
	}
	return "{" + q.NS + "}" + q.Local
}

// Schema is the IR for one XSD (either a top-level .xsd or an inline
// <wsdl:types><xs:schema>).
type Schema struct {
	File             string
	TargetNS         string
	ElementFormDef   string // "qualified" or "unqualified"; default "unqualified"
	AttributeFormDef string
	Imports          []SchemaImport // <xs:import namespace=... schemaLocation=...>
	Includes         []SchemaInclude

	SimpleTypes  []*SimpleType
	ComplexTypes []*ComplexType
	Elements     []*Element // top-level
	Attributes   []*Attribute
	AttrGroups   []*AttrGroup
	Groups       []*Group
}

// SchemaImport is an <xs:import> reference.
type SchemaImport struct {
	Namespace      string
	SchemaLocation string
}

// SchemaInclude is an <xs:include> reference (same namespace, chameleon).
type SchemaInclude struct {
	SchemaLocation string
}

// SimpleType is an <xs:simpleType>. The body is either a restriction with
// facets (most common in ONVIF: enumerations + length/pattern bounds) or a
// list/union (rare).
type SimpleType struct {
	Name        string // empty for anonymous inline
	Doc         string
	Base        QName // restriction base; zero if list/union
	Enumeration []string
	// Other facets captured raw; enough to drive Go validation later.
	Pattern      string
	MinLength    string
	MaxLength    string
	MinInclusive string
	MaxInclusive string
	Length       string
	ListItemType QName   // set if the body is <xs:list>
	UnionMembers []QName // set if the body is <xs:union memberTypes=...>
}

// ComplexType is an <xs:complexType>. The content model is captured in
// Content as a typed tree (sequence/choice/all/any) or nil for empty
// complexTypes. ComplexContent extension/restriction is recorded in
// ContentModel. Attributes are listed in Attributes.
type ComplexType struct {
	Name         string // empty for anonymous inline types
	Doc          string
	Abstract     bool
	Mixed        bool
	ContentModel ContentModel // zero value = no content model (e.g. complexType w/ sequence only)
	Content      Particle     // sequence/choice/all/any tree; nil if none
	Attributes   []*Attribute
	AnyAttribute []AnyAttribute
}

// ContentModel describes complexContent/simpleContent.
type ContentModel struct {
	Kind        ContentModelKind
	Base        QName // base type for extension/restriction
	OpenContent bool  // true if the model is implicitly open (xs:any) — informational
}

// ContentModelKind enumerates content models.
type ContentModelKind int

const (
	ContentModelNone ContentModelKind = iota
	ContentModelComplexContentExtension
	ContentModelComplexContentRestriction
	ContentModelSimpleContentExtension
	ContentModelSimpleContentRestriction
)

// Particle is a content-model node. Exactly one of the fields is meaningful
// depending on Kind.
type Particle struct {
	Kind    ParticleKind
	Min     string
	Max     string
	Element *ElementRef // for Kind == ParticleElement
	Group   *GroupRef   // for Kind == ParticleGroup
	Choice  *Choice     // for Kind == ParticleChoice
	Seq     []Particle  // for Kind == ParticleSequence / ParticleAll (ordered list)
	Any     *Any        // for Kind == ParticleAny
}

// ParticleKind enumerates content-model particles.
type ParticleKind int

const (
	ParticleEmpty ParticleKind = iota
	ParticleSequence
	ParticleChoice
	ParticleAll
	ParticleAny
	ParticleElement
	ParticleGroup
)

// ElementRef is a reference to a (local or global) element inside a sequence.
type ElementRef struct {
	Name              string
	Type              QName
	Ref               QName // set when the particle references a global <xs:element>
	MinOccurs         string
	MaxOccurs         string
	Nillable          bool
	SubstitutionGroup QName
}

// Choice groups particles; ONVIF mostly uses single-level choices but the
// parser records nesting for fidelity.
type Choice struct {
	Min  string
	Max  string
	Body []Particle
}

// Any captures <xs:any> — wildcard element.
type Any struct {
	Min             string
	Max             string
	Namespace       string // ##any, ##other, or a space-separated list
	ProcessContents string
}

// AnyAttribute captures <xs:anyAttribute>.
type AnyAttribute struct {
	Namespace       string
	ProcessContents string
}

// Attribute is an <xs:attribute> (global or local).
type Attribute struct {
	Name       string
	Type       QName
	Doc        string
	Use        string // optional / required / prohibited
	Default    string
	Fixed      string
	SimpleType *SimpleType // anonymous inline simple type
}

// AttrGroup is an <xs:attributeGroup>; its members are resolved into a flat
// attribute list at dump time so the generator doesn't need to chase refs.
type AttrGroup struct {
	Name         string
	Doc          string
	Attributes   []*Attribute
	AnyAttribute []AnyAttribute
}

// Group is an <xs:group>; its body is one Particle (typically a sequence).
type Group struct {
	Name string
	Doc  string
	Body Particle
}

// GroupRef references a global group within a content model.
type GroupRef struct {
	Ref       QName
	MinOccurs string
	MaxOccurs string
}

// Element is a global <xs:element>. Inline complex/simple types are recorded
// via InlineComplex / InlineSimple when the element doesn't reference a named
// type.
type Element struct {
	Name              string
	Doc               string
	Type              QName
	InlineComplex     *ComplexType
	InlineSimple      *SimpleType
	Nillable          bool
	SubstitutionGroup QName
	Abstract          bool
}

// WSDL is a parsed WSDL document.
type WSDL struct {
	File      string
	TargetNS  string
	Imports   []WSDLImport
	Types     []*Schema // inline <wsdl:types><xs:schema>...
	Messages  []*Message
	PortTypes []*PortType
	Bindings  []*Binding
	Services  []*Service
}

// WSDLImport is a <wsdl:import location=... namespace=...>.
type WSDLImport struct {
	Namespace string
	Location  string
}

// Message is a <wsdl:message>. Parts are element-typed in ONVIF (rpc/literal
// is not used).
type Message struct {
	Name  string
	Parts []MessagePart
}

// MessagePart is a <wsdl:part>.
type MessagePart struct {
	Name    string
	Element QName
	Type    QName
}

// PortType is a <wsdl:portType>. Operations are the public ONVIF API.
type PortType struct {
	Name       string
	Doc        string
	Operations []*Operation
}

// Operation is a <wsdl:operation>. Input/Output/Faults reference messages by
// QName.
type Operation struct {
	Name           string
	Doc            string
	ParameterOrder string
	Input          *OperationMsg
	Output         *OperationMsg
	Faults         []OperationFault
}

// OperationMsg references a message for input or output.
type OperationMsg struct {
	Message QName
	Name    string // optional part name
}

// OperationFault references a fault message.
type OperationFault struct {
	Name    string
	Message QName
}

// Binding is a <wsdl:binding>. Only the SOAP/HTTP bits relevant to ONVIF are
// captured: the SOAP action per operation and body use/style.
type Binding struct {
	Name       string
	PortType   QName
	Type       string // "soap12" or "soap11"
	Style      string // "document" or "rpc"
	Transport  string
	Operations []BindingOperation
}

// BindingOperation maps a portType operation to a SOAP action.
type BindingOperation struct {
	Name       string
	SOAPAction string
	Style      string
	Input      *BindingMsg
	Output     *BindingMsg
	Faults     []BindingFault
}

// BindingMsg records the SOAP body use ("literal") for a binding op input/output.
type BindingMsg struct {
	Use string
}

// BindingFault records the SOAP body use for a fault.
type BindingFault struct {
	Name string
	Use  string
}

// Service is a <wsdl:service>.
type Service struct {
	Name  string
	Doc   string
	Ports []ServicePort
}

// ServicePort is a <wsdl:port>.
type ServicePort struct {
	Name    string
	Binding QName
	Address string // soap:address location
}

// SortMembers orders schema members deterministically by name, preserving a
// stable iteration order for the golden dump and generated code.
func (s *Schema) SortMembers() {
	less := func(a, b string) bool { return a < b }
	sort.SliceStable(s.SimpleTypes, func(i, j int) bool { return less(s.SimpleTypes[i].Name, s.SimpleTypes[j].Name) })
	sort.SliceStable(s.ComplexTypes, func(i, j int) bool { return less(s.ComplexTypes[i].Name, s.ComplexTypes[j].Name) })
	sort.SliceStable(s.Elements, func(i, j int) bool { return less(s.Elements[i].Name, s.Elements[j].Name) })
	sort.SliceStable(s.Attributes, func(i, j int) bool { return less(s.Attributes[i].Name, s.Attributes[j].Name) })
	sort.SliceStable(s.AttrGroups, func(i, j int) bool { return less(s.AttrGroups[i].Name, s.AttrGroups[j].Name) })
	sort.SliceStable(s.Groups, func(i, j int) bool { return less(s.Groups[i].Name, s.Groups[j].Name) })
}

// Module uniquely identifies a parsed top-level artifact (WSDL or XSD) by its
// catalog path.
type Module struct {
	Path     string // catalog-relative path
	Kind     ModuleKind
	TargetNS string
	WSDL     *WSDL
	Schema   *Schema
}

// ModuleKind distinguishes WSDL from standalone XSD modules.
type ModuleKind int

const (
	ModuleWSDL ModuleKind = iota
	ModuleSchema
)
