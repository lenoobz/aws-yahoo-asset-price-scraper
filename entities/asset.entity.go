package entities

type Asset struct {
	Ticker   string `json:"ticker,omitempty"`
	Currency string `json:"currency,omitempty"`
}
