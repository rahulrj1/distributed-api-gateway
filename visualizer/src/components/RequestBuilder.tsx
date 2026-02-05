import { useState } from 'react';
import { RequestConfig } from '../types';

interface RequestBuilderProps {
  onSend: (config: RequestConfig, token: string) => void;
  isLoading: boolean;
}

const METHODS = ['GET', 'POST', 'PUT', 'DELETE'] as const;
const EXAMPLE_PATHS = [
  '/service-a/hello',
  '/service-b/hello',
  '/service-c/hello',
  '/service-a/api/data',
];

export function RequestBuilder({ onSend, isLoading }: RequestBuilderProps) {
  const [method, setMethod] = useState<RequestConfig['method']>('GET');
  const [path, setPath] = useState('/service-a/hello');
  const [token, setToken] = useState('');
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [customHeaders, setCustomHeaders] = useState('');
  const [body, setBody] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    const headers: Record<string, string> = {};
    if (customHeaders.trim()) {
      customHeaders.split('\n').forEach(line => {
        const [key, ...valueParts] = line.split(':');
        if (key && valueParts.length) {
          headers[key.trim()] = valueParts.join(':').trim();
        }
      });
    }

    onSend({ method, path, headers, body: body || undefined }, token);
  };

  return (
    <form onSubmit={handleSubmit} className="bg-gray-800 rounded-lg p-6 space-y-4">
      <h2 className="text-xl font-semibold text-white mb-4">Request Builder</h2>
      
      {/* Method & Path */}
      <div className="flex gap-2">
        <select
          value={method}
          onChange={(e) => setMethod(e.target.value as RequestConfig['method'])}
          className="bg-gray-700 text-white px-3 py-2 rounded-lg border border-gray-600 focus:border-blue-500 focus:outline-none"
        >
          {METHODS.map(m => (
            <option key={m} value={m}>{m}</option>
          ))}
        </select>
        
        <input
          type="text"
          value={path}
          onChange={(e) => setPath(e.target.value)}
          placeholder="/service-a/hello"
          className="flex-1 bg-gray-700 text-white px-4 py-2 rounded-lg border border-gray-600 focus:border-blue-500 focus:outline-none"
        />
      </div>

      {/* Quick path buttons */}
      <div className="flex flex-wrap gap-2">
        {EXAMPLE_PATHS.map(p => (
          <button
            key={p}
            type="button"
            onClick={() => setPath(p)}
            className={`px-3 py-1 text-sm rounded-full transition-colors ${
              path === p 
                ? 'bg-blue-600 text-white' 
                : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
            }`}
          >
            {p}
          </button>
        ))}
      </div>

      {/* JWT Token */}
      <div>
        <label className="block text-gray-400 text-sm mb-1">JWT Token</label>
        <textarea
          value={token}
          onChange={(e) => setToken(e.target.value)}
          placeholder="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
          rows={2}
          className="w-full bg-gray-700 text-white px-4 py-2 rounded-lg border border-gray-600 focus:border-blue-500 focus:outline-none font-mono text-sm"
        />
        <p className="text-gray-500 text-xs mt-1">
          Generate with: <code className="bg-gray-700 px-1 rounded">go run scripts/generate_jwt.go -sub user -client_id test -exp 1h</code>
        </p>
      </div>

      {/* Advanced toggle */}
      <button
        type="button"
        onClick={() => setShowAdvanced(!showAdvanced)}
        className="text-blue-400 text-sm hover:text-blue-300"
      >
        {showAdvanced ? '▼ Hide Advanced' : '▶ Show Advanced'}
      </button>

      {/* Advanced options */}
      {showAdvanced && (
        <div className="space-y-3 pl-4 border-l-2 border-gray-700">
          <div>
            <label className="block text-gray-400 text-sm mb-1">Custom Headers (one per line)</label>
            <textarea
              value={customHeaders}
              onChange={(e) => setCustomHeaders(e.target.value)}
              placeholder="X-Custom-Header: value"
              rows={2}
              className="w-full bg-gray-700 text-white px-4 py-2 rounded-lg border border-gray-600 focus:border-blue-500 focus:outline-none font-mono text-sm"
            />
          </div>
          
          {(method === 'POST' || method === 'PUT') && (
            <div>
              <label className="block text-gray-400 text-sm mb-1">Request Body (JSON)</label>
              <textarea
                value={body}
                onChange={(e) => setBody(e.target.value)}
                placeholder='{"key": "value"}'
                rows={3}
                className="w-full bg-gray-700 text-white px-4 py-2 rounded-lg border border-gray-600 focus:border-blue-500 focus:outline-none font-mono text-sm"
              />
            </div>
          )}
        </div>
      )}

      {/* Submit button */}
      <button
        type="submit"
        disabled={isLoading || !token.trim()}
        className={`w-full py-3 rounded-lg font-semibold transition-all ${
          isLoading || !token.trim()
            ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
            : 'bg-blue-600 text-white hover:bg-blue-500 active:scale-[0.98]'
        }`}
      >
        {isLoading ? (
          <span className="flex items-center justify-center gap-2">
            <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
            Processing...
          </span>
        ) : (
          'Send Request & Visualize Pipeline'
        )}
      </button>
    </form>
  );
}
