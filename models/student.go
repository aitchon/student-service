package models

type Student struct {
	ID    int    `json:"id" binding:"omitempty"`
	Name  string `json:"name" binding:"required"`
	Grade int    `json:"grade" validate:"required,min=0,max=100"`
}
