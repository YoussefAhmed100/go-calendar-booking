package service

// Calendar is the public representation of a Google calendar. It never contains credentials.
type Calendar struct {
	ID         string `json:"id"`
	Summary    string `json:"summary"`
	Primary    bool   `json:"primary"`
	AccessRole string `json:"accessRole"`
	TimeZone   string `json:"timeZone"`
}
