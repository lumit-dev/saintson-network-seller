package models

type DeleteResponse struct{
	Response struct {
		IsDeleted bool `json:"isDeleted"`
	}
}