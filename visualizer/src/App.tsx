import { useState, useCallback } from 'react';
import { RequestBuilder } from './components/RequestBuilder';
import { PipelineVisualization } from './components/PipelineVisualization';
import { useTraceWebSocket } from './hooks/useTraceWebSocket';
import { RequestConfig } from './types';

// Generate a unique trace ID
function generateTraceId(): string {
  return Array.from(crypto.getRandomValues(new Uint8Array(16)))
    .map(b => b.toString(16).padStart(2, '0'))
    .join('');
}

export default function App() {
  const [gatewayUrl, setGatewayUrl] = useState('http://localhost:5000');
  const [isLoading, setIsLoading] = useState(false);
  const [response, setResponse] = useState<{ status: number; body: string } | null>(null);

  const { steps, traceId, isConnected, error, connect, reset } = useTraceWebSocket(gatewayUrl);

  const handleSendRequest = useCallback(async (config: RequestConfig, token: string) => {
    setIsLoading(true);
    setResponse(null);
    
    // Generate new trace ID and connect to WebSocket
    const newTraceId = generateTraceId();
    connect(newTraceId);

    // Wait a bit for WebSocket to connect
    await new Promise(resolve => setTimeout(resolve, 300));

    try {
      // Build request
      const headers: HeadersInit = {
        'Authorization': `Bearer ${token}`,
        'X-Trace-ID': newTraceId,
        ...config.headers,
      };

      if (config.body) {
        headers['Content-Type'] = 'application/json';
      }

      const res = await fetch(`${gatewayUrl}${config.path}`, {
        method: config.method,
        headers,
        body: config.body,
      });

      const body = await res.text();
      setResponse({ status: res.status, body });
    } catch (err) {
      setResponse({ 
        status: 0, 
        body: `Request failed: ${err instanceof Error ? err.message : 'Unknown error'}` 
      });
    } finally {
      setIsLoading(false);
    }
  }, [gatewayUrl, connect]);

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      {/* Header */}
      <header className="bg-gray-800 border-b border-gray-700">
        <div className="max-w-7xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold">🔀 API Gateway Pipeline Visualizer</h1>
              <p className="text-gray-400 text-sm">Watch requests flow through the middleware chain in real-time</p>
            </div>
            <div className="flex items-center gap-3">
              <label className="text-gray-400 text-sm">Gateway:</label>
              <input
                type="text"
                value={gatewayUrl}
                onChange={(e) => setGatewayUrl(e.target.value)}
                className="bg-gray-700 text-white px-3 py-1.5 rounded border border-gray-600 focus:border-blue-500 focus:outline-none text-sm w-48"
              />
              <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-gray-500'}`} 
                   title={isConnected ? 'WebSocket connected' : 'WebSocket disconnected'} />
            </div>
          </div>
        </div>
      </header>

      {/* Main content */}
      <main className="max-w-7xl mx-auto px-4 py-6">
        <div className="grid lg:grid-cols-2 gap-6">
          {/* Left: Request Builder */}
          <div className="space-y-6">
            <RequestBuilder onSend={handleSendRequest} isLoading={isLoading} />
            
            {/* Response */}
            {response && (
              <div className="bg-gray-800 rounded-lg p-6">
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-xl font-semibold">Response</h2>
                  <span className={`px-3 py-1 rounded-full text-sm font-medium ${
                    response.status >= 200 && response.status < 300 ? 'bg-green-900 text-green-300' :
                    response.status >= 400 && response.status < 500 ? 'bg-yellow-900 text-yellow-300' :
                    response.status >= 500 ? 'bg-red-900 text-red-300' :
                    'bg-gray-700 text-gray-300'
                  }`}>
                    {response.status || 'Error'}
                  </span>
                </div>
                <pre className="bg-gray-900 rounded p-4 overflow-x-auto text-sm font-mono text-gray-300 max-h-64 overflow-y-auto">
                  {response.body || '(empty response)'}
                </pre>
              </div>
            )}
          </div>

          {/* Right: Pipeline Visualization */}
          <div className="space-y-6">
            <PipelineVisualization steps={steps} traceId={traceId} />
            
            {/* Error display */}
            {error && (
              <div className="bg-red-900/30 border border-red-700 rounded-lg p-4 text-red-300">
                <strong>Error:</strong> {error}
              </div>
            )}

            {/* Reset button */}
            {steps.length > 0 && (
              <button
                onClick={reset}
                className="w-full py-2 bg-gray-700 hover:bg-gray-600 rounded-lg transition-colors"
              >
                Reset Visualization
              </button>
            )}
          </div>
        </div>

        {/* Legend */}
        <div className="mt-8 p-4 bg-gray-800 rounded-lg">
          <h3 className="text-sm font-medium text-gray-400 mb-3">Pipeline Steps Legend</h3>
          <div className="flex flex-wrap gap-4 text-sm">
            <div className="flex items-center gap-2">
              <span className="w-4 h-4 rounded-full bg-gray-600" />
              <span className="text-gray-400">Pending</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="w-4 h-4 rounded-full bg-blue-600 animate-pulse" />
              <span className="text-gray-400">Running</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="w-4 h-4 rounded-full bg-green-600" />
              <span className="text-gray-400">Success</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="w-4 h-4 rounded-full bg-red-600" />
              <span className="text-gray-400">Failed</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="w-4 h-4 rounded-full bg-purple-600" />
              <span className="text-gray-400">Skipped</span>
            </div>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t border-gray-800 mt-8 py-4">
        <div className="max-w-7xl mx-auto px-4 text-center text-gray-500 text-sm">
          Distributed API Gateway • Pipeline Visualizer
        </div>
      </footer>
    </div>
  );
}
