package labels

var carLabels = map[string]string{}
var controlLabels = map[string]string{}

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
