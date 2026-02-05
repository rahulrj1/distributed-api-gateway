import { useState } from 'react';
import { RequestConfig } from '../types';

interface RequestBuilderProps {
  onSend: (config: RequestConfig, token: string) => void;
  isLoading: boolean;
}

const EXAMPLE_PATHS = [
  '/service-a/hello',
  '/service-b/hello',
  '/service-c/hello',
];

const JWT_COMMAND = 'go run scripts/generate_jwt.go -sub user -client_id test -exp 1h';

export function RequestBuilder({ onSend, isLoading }: RequestBuilderProps) {
  const [path, setPath] = useState('/service-a/hello');
  const [token, setToken] = useState('');
  const [copied, setCopied] = useState(false);

  const copyCommand = () => {
    navigator.clipboard.writeText(JWT_COMMAND);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSend({ method: 'GET', path, headers: {} }, token);
  };

  return (
    <form onSubmit={handleSubmit} className="bg-gray-800 rounded-lg p-6 space-y-4">
      <h2 className="text-xl font-semibold text-white mb-4">Request Builder</h2>
      
      {/* Path */}
      <div>
        <label className="block text-gray-400 text-sm mb-1">Endpoint</label>
        <input
          type="text"
          value={path}
          onChange={(e) => setPath(e.target.value)}
          placeholder="/service-a/hello"
          className="w-full bg-gray-700 text-white px-4 py-2 rounded-lg border border-gray-600 focus:border-blue-500 focus:outline-none"
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
        <input
          type="text"
          value={token}
          onChange={(e) => setToken(e.target.value)}
          placeholder="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
          className="w-full bg-gray-700 text-white px-4 py-2 rounded-lg border border-gray-600 focus:border-blue-500 focus:outline-none font-mono text-sm"
        />
        <div className="mt-2 p-3 bg-gray-700/50 rounded-lg border border-gray-600">
          <div className="flex items-center justify-between mb-1">
            <span className="text-gray-400 text-xs font-medium">Generate JWT token:</span>
            <button
              type="button"
              onClick={copyCommand}
              className="text-gray-400 hover:text-white transition-colors p-1 rounded hover:bg-gray-600"
              title="Copy command"
            >
              {copied ? (
                <svg className="w-4 h-4 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              ) : (
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                </svg>
              )}
            </button>
          </div>
          <code className="text-gray-300 text-xs font-mono block overflow-x-auto">
            {JWT_COMMAND}
          </code>
        </div>
      </div>

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
