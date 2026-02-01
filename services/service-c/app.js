/**
 * Service C - Node.js/Express Backend Service
 * 
 * Endpoints:
 * - GET /health - Health check
 * - GET /hello - Hello message
 * - POST /echo - Echo request details
 */

const express = require('express');
const app = express();

app.use(express.json());

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({
    status: 'healthy',
    service: 'service-c'
  });
});

// Hello endpoint
app.get('/hello', (req, res) => {
  res.json({
    message: 'Hello from Node.js Service C',
    request_id: req.headers['x-request-id'] || 'unknown',
    user_id: req.headers['x-user-id'] || 'unknown'
  });
});

// Echo endpoint - accepts all methods
app.all('/echo', (req, res) => {
  const headers = { ...req.headers };
  delete headers['host'];
  delete headers['content-length'];

  res.json({
    service: 'service-c',
    method: req.method,
    path: req.path,
    query_params: req.query,
    headers: headers,
    body: req.body
  });
});

// Export for testing
module.exports = app;

// Start server if run directly
if (require.main === module) {
  const port = process.env.PORT || 6002;
  app.listen(port, () => {
    console.log(`Service C running on port ${port}`);
  });
}
