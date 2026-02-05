import { useState, useCallback, useRef } from 'react';
import { TraceEvent, StepState, Step, Status } from '../types';

const STEP_LABELS: Record<Step, string> = {
  RECEIVED: 'Request Received',
  AUTH: 'Authentication',
  RATE_LIMIT: 'Rate Limiting',
  CIRCUIT: 'Circuit Breaker',
  FORWARD: 'Forward to Backend',
  RESPONSE: 'Response',
  COMPLETE: 'Complete',
};

// Order of steps in the pipeline
const STEP_ORDER: Step[] = ['RECEIVED', 'AUTH', 'RATE_LIMIT', 'CIRCUIT', 'FORWARD', 'COMPLETE'];

function createInitialSteps(): StepState[] {
  return STEP_ORDER.map(step => ({
    step,
    label: STEP_LABELS[step],
    status: 'PENDING' as Status,
  }));
}

export function useTraceWebSocket(gatewayUrl: string) {
  const [steps, setSteps] = useState<StepState[]>([]);
  const [traceId, setTraceId] = useState<string>();
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string>();
  const wsRef = useRef<WebSocket | null>(null);

  const connect = useCallback((newTraceId: string) => {
    // Close existing connection
    if (wsRef.current) {
      wsRef.current.close();
    }

    setTraceId(newTraceId);
    setSteps(createInitialSteps());
    setError(undefined);

    // Convert http:// to ws://
    const wsUrl = gatewayUrl.replace(/^http/, 'ws') + `/ws/trace/${newTraceId}`;
    console.log('Connecting to WebSocket:', wsUrl);

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log('WebSocket connected');
      setIsConnected(true);
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        
        // Handle subscription confirmation
        if (data.type === 'subscribed') {
          console.log('Subscribed to trace:', data.trace_id);
          return;
        }

        // Handle timeout/error
        if (data.type === 'timeout' || data.type === 'error') {
          setError(data.message || data.error);
          return;
        }

        // Handle trace event
        const traceEvent = data as TraceEvent;
        console.log('Trace event:', traceEvent);

        setSteps(prev => {
          const newSteps = [...prev];
          const stepIndex = newSteps.findIndex(s => s.step === traceEvent.step);
          
          if (stepIndex !== -1) {
            newSteps[stepIndex] = {
              ...newSteps[stepIndex],
              status: traceEvent.status,
              duration: traceEvent.duration_us,
              error: traceEvent.error,
              details: traceEvent.details,
            };
          }
          
          return newSteps;
        });
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e);
      }
    };

    ws.onerror = (event) => {
      console.error('WebSocket error:', event);
      setError('WebSocket connection error');
    };

    ws.onclose = () => {
      console.log('WebSocket closed');
      setIsConnected(false);
    };
  }, [gatewayUrl]);

  const disconnect = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setIsConnected(false);
  }, []);

  const reset = useCallback(() => {
    disconnect();
    setSteps([]);
    setTraceId(undefined);
    setError(undefined);
  }, [disconnect]);

  return {
    steps,
    traceId,
    isConnected,
    error,
    connect,
    disconnect,
    reset,
  };
}
