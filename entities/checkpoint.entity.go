package entities

type CheckPoint struct {
	PageSize  int64 `json:"size,omitempty"`
	PageIndex int64 `json:"index,omitempty"`
}
