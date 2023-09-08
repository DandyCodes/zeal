package models

type Item struct {
	Name  string
	Price float32
}

type Menu struct {
	ID    int
	Items []Item
}
