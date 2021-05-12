package entities

type Security struct {
	Ticker   string `json:"ticker,omitempty"`
	Currency string `json:"currency,omitempty"`
}
