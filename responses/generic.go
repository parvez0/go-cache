package responses

type GenericResponse struct {
	Success bool `json:"success"`
	Message string `json:"message"`
	Data map[string]string `json:"data"`
}
