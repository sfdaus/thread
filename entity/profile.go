package entity

type Profile struct {
	PublicID    string      `json:"public_id"`
	Name        string      `json:"name"`
	NameAlias   string      `json:"name_alias"`
	Avatar      string      `json:"avatar"`
	Institution Institution `json:"institution"`
}

type SimpleProfile struct {
	Name      string `json:"name"`
	NameAlias string `json:"name_alias"`
	Avatar    string `json:"avatar"`
}
