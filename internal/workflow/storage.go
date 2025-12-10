package workflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

var (
	workflowFile = filepath.Join("data", "workflow.json")
	mu           sync.RWMutex
)

// LoadCurrentWorkflow loads the current workflow from file or returns default
func LoadCurrentWorkflow() (*Workflow, error) {
	mu.RLock()
	defer mu.RUnlock()

	data, err := os.ReadFile(workflowFile)
	if err != nil {
		if os.IsNotExist(err) {
			return CreateDefaultConversationWorkflow(), nil
		}
		return nil, err
	}

	var wf Workflow
	if err := json.Unmarshal(data, &wf); err != nil {
		return nil, err
	}
	return &wf, nil
}

// SaveWorkflow saves the workflow to file
func SaveWorkflow(wf *Workflow) error {
	mu.Lock()
	defer mu.Unlock()

	data, err := json.MarshalIndent(wf, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(workflowFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(workflowFile, data, 0644)
}
