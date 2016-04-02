package tpt

import "net/url"

type UserAccountDetails struct {
	ID              string `json:"id"`
	Status          string `json:"status"`
	Discretionary   bool   `json:"discretionary"`
	NonProfessional bool   `json:"non_professional"`
	AccountType     string `json:"account_type"`
}
type UserAccountApplicant struct {
	FirstName          string  `json:"first_name"`
	MiddleName         string  `json:"middle_name"`
	LastName           string  `json:"last_name"`
	Email              string  `json:"email"`
	Birthday           string  `json:"birthday"`
	SsnLast4           string  `json:"ssn_last_4"`
	BirthCountry       string  `json:"birth_country"`
	CitizenshipCountry string  `json:"citizenship_country"`
	Mobile             string  `json:"mobile"`
	MobileDevice       string  `json:"mobile_device"`
	Address            Address `json:"address"`
}
type Address struct {
	Line1      *string `json:"line_1,omitempty"`
	Line2      *string `json:"line_2,omitempty"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	PostalCode string  `json:"postal_code"`
	Country    string  `json:"country"`
}

type NewsResponse struct {
	Items []NewsItem `json:"news"`
}

type NewsItem struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Summary string `json:"summary"`
}

func (c *Client) GetNews(symbol string) (*NewsResponse, error) {
	resp := &NewsResponse{}
	err := c.New("/v1/news").
		Query(&url.Values{"symbol": {symbol}}).
		DecodeInto(resp)
	return resp, err
}
