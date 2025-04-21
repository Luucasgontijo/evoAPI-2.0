package middleware

import (
    "bytes"
    "io"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

// Logger middleware para logar todas as requisições com ID de correlação
func RequestLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Gera um ID de correlação único para cada requisição
        correlationID := uuid.New().String()
        c.Set("correlationID", correlationID)
        c.Header("X-Correlation-ID", correlationID)
        
        // Captura o início da requisição
        startTime := time.Now()
        
        // Captura o corpo da requisição
        var requestBody []byte
        if c.Request.Body != nil {
            requestBody, _ = io.ReadAll(c.Request.Body)
            c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
        }
        
        // Prossegue para o próximo middleware/handler
        c.Next()
        
        // Após o processamento, registra os detalhes da requisição
        duration := time.Since(startTime)
        status := c.Writer.Status()
        
        // Filtra informações sensíveis do log
        method := c.Request.Method
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery
        ip := c.ClientIP()
        
        // Registra a requisição no formato: [CorrelationID] Method Path StatusCode Duration ClientIP
        logEntry := gin.LogFormatter{
            TimeStamp:    time.Now(),
            StatusCode:   status,
            Latency:      duration,
            ClientIP:     ip,
            Method:       method,
            Path:         path,
            ErrorMessage: c.Errors.String(),
            BodySize:     c.Writer.Size(),
        }
        
        // Em um ambiente de produção, você usaria um logger estruturado (como zap ou logrus)
        // Aqui estamos usando o logger padrão do Gin para simplicidade
        gin.DefaultWriter.Write([]byte("[" + correlationID + "] " + logEntry.ClientIP + " " + 
            logEntry.Method + " " + logEntry.Path + 
            " " + query + " " + 
            " " + string(status) + " " + 
            duration.String() + "\n"))
    }
}