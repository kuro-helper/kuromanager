package erogssync

// Erogs dump.sql 回傳的品牌資料。
type DumpBrand struct {
	ID       int        `json:"id"`
	Name     string     `json:"name"`
	Disband  bool       `json:"disband"`
	Gamelist []DumpGame `json:"gamelist"`
}

// Erogs dump.sql 回傳的遊戲資料。
type DumpGame struct {
	ID           int    `json:"id"`
	BrandErogsID int    `json:"brandErogsId"`
	Name         string `json:"name"`
	Category     string `json:"category"`
	Image        string `json:"image"`
}
