package ginmodel

type DeleteDatabaseItemRequest struct {
	Key string `json:"key"`
}

type UpdateDatabaseItemRequest struct {
	Key  string `json:"key"`
	Val string `json:"item"`
}