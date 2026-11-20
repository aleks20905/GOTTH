package templates

// ComputerRow is one agent's status as shown on the computers page.
// Offline rows carry "—" in the value fields.
type ComputerRow struct {
	URL     string
	Name    string
	Version string
	OS      string
	Uptime  string
	Online  bool
}
