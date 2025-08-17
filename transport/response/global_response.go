package response

// Meta Response
type MetaRes struct {
	Page       int64 `json:"page"`
	PerPage    int64 `json:"per_page"`
	TotalData  int64 `json:"total_data"`
	TotalPages int64 `json:"total_pages"`
}
