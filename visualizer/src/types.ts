// Pipeline step types matching backend
export type Step = 
  | 'RECEIVED' 
  | 'AUTH' 
  | 'RATE_LIMIT' 
  | 'CIRCUIT' 
  | 'FORWARD' 
  | 'RESPONSE' 
  | 'COMPLETE';

export type Status = 'PENDING' | 'RUNNING' | 'SUCCESS' | 'FAILED' | 'SKIPPED';

// Trace event from WebSocket
export interface TraceEvent {
  trace_id: string;
  step: Step;
  status: Status;
  timestamp: string;
  duration_us?: number;
  error?: string;
  details?: Record<string, unknown>;
}

// Request configuration for the builder
export interface RequestConfig {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE';
  path: string;
  headers: Record<string, string>;
  body?: string;
}

// Pipeline step state for visualization
export interface StepState {
  step: Step;
  label: string;
  status: Status;
  duration?: number; // milliseconds
  error?: string;
  details?: Record<string, unknown>;
}

// WebSocket message types
export interface WSSubscribed {
  type: 'subscribed';
  trace_id: string;
  channel: string;
}

export interface WSError {
  type?: 'error' | 'timeout';
  error?: string;
  message?: string;
}
