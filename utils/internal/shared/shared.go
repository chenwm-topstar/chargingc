package shared

var (
	// PtrServiceName is a shared flags for global use.
	// User should use GetServiceName() instead.
	PtrServiceName *string
	PtrConsulAddr  *string
)

// GetServiceName returns special `service` flag always included in this library.
func GetServiceName() string {
	if PtrServiceName == nil {
		return ""
	}
	return *PtrServiceName
}

func GetConsulAddress() string {
	if PtrConsulAddr == nil {
		return ""
	}
	return *PtrConsulAddr
}
