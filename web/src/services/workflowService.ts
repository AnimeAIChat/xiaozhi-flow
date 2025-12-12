import axios from 'axios';

export interface Capability {
  id: string;
  type: string;
  name: string;
  description: string;
  input_schema: any;
  output_schema: any;
  config_schema: any;
}

export interface WorkflowNode {
  id: string;
  name: string;
  type: string;
  plugin: string;
  description: string;
  inputs: any[];
  outputs: any[];
  position: { x: number; y: number };
  config?: any;
}

export interface WorkflowEdge {
  id: string;
  from: string;
  to: string;
}

export interface Workflow {
  id: string;
  name: string;
  description: string;
  nodes: WorkflowNode[];
  edges: WorkflowEdge[];
  config: any;
}

const API_BASE_URL = '/api/v1/workflow';

export const workflowService = {
  getCapabilities: async (): Promise<Capability[]> => {
    const response = await axios.get(`${API_BASE_URL}/capabilities`);
    return response.data.data;
  },

  getCurrentWorkflow: async (): Promise<Workflow> => {
    const response = await axios.get(`${API_BASE_URL}/current`);
    return response.data.data;
  },

  saveWorkflow: async (workflow: Workflow): Promise<Workflow> => {
    const response = await axios.post(API_BASE_URL, workflow);
    return response.data.data;
  },
};
