// Package iiss carries the IISS form structs (Class 0/I, Class II/III,
// Class IV-P) and their renderers. Per docs/pass-2/api-surface.md §
// Package boundaries, iiss/ is split out from worlds/ so the renderer
// concerns are separately replaceable. iiss/ does not import worlds/;
// the SystemForms aggregate is the boundary type.
package iiss

// FormHeader is the header common to all three IISS forms. Per anti-pattern
// A.5, sector and location are separate fields, not a delimited string.
type FormHeader struct {
	SystemName    string
	Sector        string
	Location      string
	InitialSurvey string
	LastUpdated   string
	IISSDesig     string
}
