package labels

// Friendly names. The car map is intentionally small for now — it covers the
// author's primary car (f4) and the control names confirmed this session.
// Additional car slugs fall back to their technical id until the author
// confirms the full friendly-name list.
var carLabels = map[string]string{
	"f4": "F4",
}

var controlLabels = map[string]string{
	"LowFuelAccept": "Low Fuel Accept",
}

// Car returns the friendly name for a car id, falling back to the id itself.
func Car(id string) string {
	if s, ok := carLabels[id]; ok {
		return s
	}
	return id
}

// Control returns the friendly name for a control id, falling back to the id.
func Control(id string) string {
	if s, ok := controlLabels[id]; ok {
		return s
	}
	return id
}
