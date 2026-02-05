import { motion, AnimatePresence } from 'framer-motion';
import { StepState, Status } from '../types';

interface PipelineVisualizationProps {
  steps: StepState[];
  traceId?: string;
}

const STATUS_COLORS: Record<Status, string> = {
  PENDING: 'bg-gray-600 border-gray-500',
  RUNNING: 'bg-blue-600 border-blue-400 animate-pulse',
  SUCCESS: 'bg-green-600 border-green-400',
  FAILED: 'bg-red-600 border-red-400',
  SKIPPED: 'bg-purple-600 border-purple-400',
};

const STATUS_ICONS: Record<Status, string> = {
  PENDING: '○',
  RUNNING: '◐',
  SUCCESS: '✓',
  FAILED: '✗',
  SKIPPED: '⊘',
};

const STEP_ICONS: Record<string, string> = {
  RECEIVED: '📥',
  AUTH: '🔐',
  RATE_LIMIT: '⏱️',
  CIRCUIT: '⚡',
  FORWARD: '🔀',
  RESPONSE: '📤',
  COMPLETE: '✅',
};

function formatDuration(us?: number): string {
  if (!us) return '';
  if (us < 1000) return `${us}µs`;
  if (us < 1000000) return `${(us / 1000).toFixed(1)}ms`;
  return `${(us / 1000000).toFixed(2)}s`;
}

function StepNode({ step, index }: { step: StepState; index: number }) {
  const isLast = step.step === 'COMPLETE';
  
  return (
    <motion.div
      initial={{ opacity: 0, x: -20 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ delay: index * 0.1 }}
      className="flex items-start gap-4"
    >
      {/* Step indicator */}
      <div className="flex flex-col items-center">
        <motion.div
          className={`w-12 h-12 rounded-full border-2 flex items-center justify-center text-xl
            ${STATUS_COLORS[step.status]}`}
          animate={step.status === 'RUNNING' ? { scale: [1, 1.1, 1] } : {}}
          transition={{ repeat: Infinity, duration: 1 }}
        >
          {STEP_ICONS[step.step] || STATUS_ICONS[step.status]}
        </motion.div>
        
        {/* Connector line */}
        {!isLast && (
          <motion.div
            className={`w-0.5 h-8 mt-2 ${
              step.status === 'SUCCESS' ? 'bg-green-500' : 
              step.status === 'FAILED' ? 'bg-red-500' : 'bg-gray-600'
            }`}
            initial={{ height: 0 }}
            animate={{ height: 32 }}
            transition={{ delay: 0.2 }}
          />
        )}
      </div>

      {/* Step content */}
      <div className="flex-1 pb-6">
        <div className="flex items-center gap-3">
          <h3 className="text-white font-medium">{step.label}</h3>
          <span className={`px-2 py-0.5 text-xs rounded-full ${
            step.status === 'SUCCESS' ? 'bg-green-900 text-green-300' :
            step.status === 'FAILED' ? 'bg-red-900 text-red-300' :
            step.status === 'RUNNING' ? 'bg-blue-900 text-blue-300' :
            'bg-gray-700 text-gray-400'
          }`}>
            {step.status}
          </span>
          {step.duration !== undefined && (
            <span className="text-gray-400 text-sm">
              {formatDuration(step.duration)}
            </span>
          )}
        </div>

        {/* Error message */}
        <AnimatePresence>
          {step.error && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              className="mt-2 p-3 bg-red-900/30 border border-red-700 rounded-lg text-red-300 text-sm"
            >
              {step.error}
            </motion.div>
          )}
        </AnimatePresence>

        {/* Details */}
        <AnimatePresence>
          {step.details && Object.keys(step.details).length > 0 && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              className="mt-2 p-3 bg-gray-700/50 rounded-lg text-sm"
            >
              <div className="grid grid-cols-2 gap-2">
                {Object.entries(step.details).map(([key, value]) => (
                  <div key={key}>
                    <span className="text-gray-400">{key}: </span>
                    <span className="text-gray-200 font-mono">
                      {typeof value === 'object' ? JSON.stringify(value) : String(value)}
                    </span>
                  </div>
                ))}
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </motion.div>
  );
}

export function PipelineVisualization({ steps, traceId }: PipelineVisualizationProps) {
  if (steps.length === 0) {
    return (
      <div className="bg-gray-800 rounded-lg p-6">
        <h2 className="text-xl font-semibold text-white mb-4">Pipeline Visualization</h2>
        <div className="text-center py-12 text-gray-400">
          <div className="text-4xl mb-4">🔍</div>
          <p>Send a request to see the pipeline in action</p>
          <p className="text-sm mt-2">
            Watch each middleware step execute in real-time
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-gray-800 rounded-lg p-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-white">Pipeline Visualization</h2>
        {traceId && (
          <span className="text-xs font-mono text-gray-400 bg-gray-700 px-2 py-1 rounded">
            Trace: {traceId.slice(0, 8)}...
          </span>
        )}
      </div>

      <div className="space-y-0">
        {steps.map((step, index) => (
          <StepNode key={step.step} step={step} index={index} />
        ))}
      </div>

      {/* Latency summary */}
      {steps.some(s => s.duration !== undefined) && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="mt-6 pt-4 border-t border-gray-700"
        >
          <h4 className="text-sm text-gray-400 mb-2">Latency Breakdown</h4>
          <div className="flex gap-2 flex-wrap">
            {steps
              .filter(s => s.duration !== undefined)
              .map(s => (
                <div key={s.step} className="bg-gray-700 rounded px-3 py-1 text-sm">
                  <span className="text-gray-400">{s.label}: </span>
                  <span className="text-white font-mono">{formatDuration(s.duration)}</span>
                </div>
              ))}
          </div>
        </motion.div>
      )}
    </div>
  );
}
