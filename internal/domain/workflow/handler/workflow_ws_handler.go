package handler

import (
	"net/http"
	"net/url"

	"auth-perm/internal/controller/util"
	"auth-perm/internal/domain/workflow/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WorkflowWSHandler struct {
	svc      *service.WorkflowService
	upgrader websocket.Upgrader
}

func NewWorkflowWSHandler(svc *service.WorkflowService) *WorkflowWSHandler {
	return &WorkflowWSHandler{
		svc: svc,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return false
				}
				originURL, err := url.Parse(origin)
				if err != nil {
					return false
				}
				if originURL.Hostname() == "localhost" || originURL.Hostname() == "127.0.0.1" {
					return true
				}
				return originURL.Hostname() == r.URL.Hostname()
			},
		},
	}
}

func (h *WorkflowWSHandler) HandleWS(c *gin.Context) {
	runID := c.Param("runId")
	if runID == "" {
		c.JSON(400, gin.H{"error": "run_id required"})
		return
	}

	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		c.JSON(400, gin.H{"error": "tenant_id required"})
		return
	}

	accountID, err := util.GetAccountID(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	run, err := h.svc.GetRun(runID, tenantID)
	if err != nil {
		c.JSON(403, gin.H{"error": "access denied"})
		return
	}
	if run.AccountID != accountID {
		c.JSON(403, gin.H{"error": "access denied"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	hub := h.svc.GetWSHub()
	client := hub.RegisterClient(runID, conn)

	go client.WritePump()
	go client.ReadPump()
}
