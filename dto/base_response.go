package dto

type BaseResponse struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error"`
	Message string      `json:"message"`
}
