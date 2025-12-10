package v1

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"xiaozhi-server-go/internal/platform/config"
	"xiaozhi-server-go/internal/platform/logging"
	"xiaozhi-server-go/internal/plugin/capability"
	"xiaozhi-server-go/internal/workflow"
)

type WorkflowService struct {
	config   *config.Config
	logger   *logging.Logger
	registry *capability.Registry
	mu       sync.RWMutex
}

func NewWorkflowService(config *config.Config, logger *logging.Logger, registry *capability.Registry) *WorkflowService {
	return &WorkflowService{
		config:   config,
		logger:   logger,
		registry: registry,
	}
}

func (s *WorkflowService) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/workflow")
	{
		group.GET("/capabilities", s.ListCapabilities)
		group.GET("/current", s.GetCurrentWorkflow)
		group.POST("", s.SaveWorkflow)
	}
}

// ListCapabilities returns all available capabilities
func (s *WorkflowService) ListCapabilities(c *gin.Context) {
	if s.registry == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "registry not initialized"})
		return
	}
	caps := s.registry.ListCapabilities()
	c.JSON(http.StatusOK, gin.H{"data": caps})
}

// GetCurrentWorkflow returns the current workflow configuration
func (s *WorkflowService) GetCurrentWorkflow(c *gin.Context) {
	wf, err := workflow.LoadCurrentWorkflow()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": wf})
}

// SaveWorkflow saves the workflow configuration
func (s *WorkflowService) SaveWorkflow(c *gin.Context) {
	var wf workflow.Workflow
	if err := c.ShouldBindJSON(&wf); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := workflow.SaveWorkflow(&wf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "workflow saved", "data": wf})
}
