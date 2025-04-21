package main

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"

	"github.com/lucasgontijo/evoAPI-2.0/evolution-middleware/middleware"
	"github.com/lucasgontijo/evoAPI-2.0/evolution-middleware/response"
	"github.com/lucasgontijo/evoAPI-2.0/evolution-middleware/websocket"

	_ "github.com/lucasgontijo/evoAPI-2.0/evolution-middleware/docs" // importa a documentação Swagger
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var evolutionClient = resty.New().
    SetBaseURL("http://localhost:8080").
    SetHeader("apikey", "a176e0c64c").
    SetTimeout(30 * time.Second)

var wsHub *websocket.Hub

func main() {
    r := gin.Default()

    // Inicializa o hub de websockets
    wsHub = websocket.NewHub()
    go wsHub.Run()

    // Middlewares globais
    r.Use(cors.Default())
    r.Use(middleware.RequestLogger())

    // Rota para documentação Swagger
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // Rota para autenticação
    r.POST("/auth/login", loginHandler)

    // Grupo de rotas protegidas por autenticação
    api := r.Group("/api")
    api.Use(middleware.AuthRequired())

    // Rota para criar instância
    api.POST("/instance/create", createInstanceHandler)

    // Rota para deletar instância
    api.DELETE("/instance/:id", deleteInstanceHandler)

    // Rota para buscar instâncias
    api.GET("/instance/fetchInstances", fetchInstancesHandler)

    // Rota para configurar webhook
    api.POST("/webhook/set/:instance", setWebhookHandler)

    // Rota para buscar configuração de webhook
    api.GET("/webhook/find/:instance", findWebhookHandler)

    // Rota para operações em lote
    api.POST("/batch/restart", batchRestartHandler)

    // Rota para websocket
    r.GET("/ws/:instance", websocketHandler)

    r.Run(":3000")
}

// @Summary Login de usuário
// @Description Autentica um usuário e retorna um token JWT
// @Accept json
// @Produce json
// @Param credentials body object true "Credenciais de login"
// @Success 200 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Router /auth/login [post]
func loginHandler(c *gin.Context) {
    var req struct {
        Username string `json:"username" binding:"required"`
        Password string `json:"password" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        resp := response.NewError(http.StatusBadRequest, "Credenciais inválidas", err.Error())
        c.JSON(resp.Code, resp)
        return
    }

    // Em um cenário real, você verificaria as credenciais em um banco de dados
    // Para este exemplo, aceitamos qualquer login com credenciais não-vazias
    if req.Username != "" && req.Password != "" {
        token, err := middleware.GenerateToken(req.Username, "user")
        if err != nil {
            resp := response.NewError(http.StatusInternalServerError, "Erro ao gerar token", err.Error())
            c.JSON(resp.Code, resp)
            return
        }

        resp := response.NewSuccess("Login bem-sucedido", gin.H{"token": token})
        c.JSON(http.StatusOK, resp)
        return
    }

    resp := response.NewError(http.StatusUnauthorized, "Credenciais inválidas", "")
    c.JSON(resp.Code, resp)
}

// Implementação dos handlers
// ... (você implementaria cada um dos handlers mencionados acima)
// Por exemplo:

func createInstanceHandler(c *gin.Context) {
    var req struct {
        InstanceName string `json:"instance_name" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        resp := response.NewError(http.StatusBadRequest, "Dados de instância inválidos", err.Error())
        c.JSON(resp.Code, resp)
        return
    }

    // Executa a requisição para a API Evolution
    evResp, err := evolutionClient.R().
        SetBody(map[string]interface{}{
            "instanceName": req.InstanceName,
            "integration":  "WHATSAPP-BAILEYS",
        }).
        Post("/instance/create")

    // Trata erros de conexão
    if err != nil {
        resp := response.NewError(http.StatusInternalServerError, "Falha ao comunicar com a API", err.Error())
        c.JSON(resp.Code, resp)
        return
    }

    // Verifica o status da resposta
    if evResp.StatusCode() != http.StatusOK && evResp.StatusCode() != http.StatusCreated {
        resp := response.NewError(evResp.StatusCode(), "Erro na API Evolution", evResp.String())
        c.JSON(resp.Code, resp)
        return
    }

    // Resposta padronizada de sucesso
    resp := response.NewSuccess("Instância criada com sucesso", gin.H{
        "instance_name": req.InstanceName,
        "api_response": evResp.String(),
    })

    c.JSON(http.StatusOK, resp)

    // Notifica clientes websocket sobre a nova instância
    wsHub.BroadcastToInstance("global", gin.H{
        "event": "instance_created",
        "data": gin.H{
            "instance_name": req.InstanceName,
            "timestamp": time.Now(),
        },
    })
}

// Implementar os demais handlers de maneira similar...

// @Summary Conectar ao websocket
// @Description Estabelece uma conexão websocket para receber atualizações em tempo real
// @Param instance path string true "Nome da instância"
// @Router /ws/{instance} [get]
func websocketHandler(c *gin.Context) {
    instance := c.Param("instance")
    // Implemente a lógica para estabelecer a conexão websocket
    // e associá-la ao hub
}