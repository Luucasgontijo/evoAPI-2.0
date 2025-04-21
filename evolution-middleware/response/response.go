package response

// StandardResponse é uma estrutura padronizada para todas as respostas da API
type StandardResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message,omitempty"`
    Error   string      `json:"error,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Code    int         `json:"code"`
}

// NewSuccess cria uma resposta de sucesso padronizada
func NewSuccess(message string, data interface{}) StandardResponse {
    return StandardResponse{
        Success: true,
        Message: message,
        Data:    data,
        Code:    200,
    }
}

// NewError cria uma resposta de erro padronizada
func NewError(code int, message string, err string) StandardResponse {
    return StandardResponse{
        Success: false,
        Message: message,
        Error:   err,
        Code:    code,
    }
}