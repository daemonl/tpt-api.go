package api

type UserAccountDetails struct {
}
type UserAccountApplicant struct {
	Address Address `json:"address"`
}
type Address struct {
	Line1      *string `json:"line_1,omitempty"`
	Line2      *string `json:"line_2,omitempty"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	PostalCode string  `json:"postal_code"`
	Country    string  `json:"country"`
}
