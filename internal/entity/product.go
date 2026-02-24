package entity

type Product struct {
	ID       int     `json:"ID"`
	Price    float64 `json:"price"`
	OldPrice float64 `json:"oldPrice"`
	Currency string  `json:"currency"`
	Validity int64   `json:"validity"`
	Name     string  `json:"name"`
}

func (p *Product) TableName() string {
	return "product"
}
