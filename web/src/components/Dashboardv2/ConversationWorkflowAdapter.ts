import {
  type Workflow,
  WorkflowEdge,
  WorkflowNode,
  workflowService,
} from '../../services/workflowService';
import type { SimpleConnection, SimpleNode } from './SimpleReteEditor';

export class ConversationWorkflowAdapter {
  async getWorkflow(workflowId: string): Promise<Workflow> {
    try {
      return await workflowService.getCurrentWorkflow();
    } catch (error) {
      console.error('Failed to get workflow:', error);
      throw error;
    }
  }

  async saveWorkflow(
    workflowId: string,
    nodes: SimpleNode[],
    connections: SimpleConnection[],
  ): Promise<void> {
    const workflow = await this.getWorkflow(workflowId);

    // Update nodes
    workflow.nodes = nodes.map((node) => ({
      id: node.id,
      name: node.data.label,
      type: 'task', // Default to task
      plugin: node.data.plugin || node.data.type, // Use plugin if available
      description: node.data.description || '',
      inputs: node.data.inputs || [],
      outputs: node.data.outputs || [],
      position: { x: node.x, y: node.y },
      config: node.data.config || {},
    }));

    // Update edges
    workflow.edges = connections.map((conn) => ({
      id: conn.id,
      from: conn.from,
      to: conn.to,
    }));

    await workflowService.saveWorkflow(workflow);
  }

  convertWorkflowToEditorNodes(workflow: Workflow): SimpleNode[] {
    return workflow.nodes.map((node) => ({
      id: node.id,
      x: node.position.x,
      y: node.position.y,
      data: {
        label: node.name,
        type: node.plugin || 'default',
        description: node.description,
        inputs: node.inputs,
        outputs: node.outputs,
        config: node.config,
        plugin: node.plugin,
        status: 'pending', // Default status
      },
    }));
  }

  convertWorkflowToEditorConnections(workflow: Workflow): SimpleConnection[] {
    return workflow.edges.map((edge) => ({
      id: edge.id,
      from: edge.from,
      to: edge.to,
      fromOutput: 'output', // Default
      toInput: 'input', // Default
    }));
  }

  async getCapabilities() {
    return await workflowService.getCapabilities();
  }
}
