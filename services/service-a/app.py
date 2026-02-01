"""
Service A - Python/Flask Backend Service

Endpoints:
- GET /health - Health check
- GET /hello - Hello message
- POST /echo - Echo request details
"""

import os
import json
from flask import Flask, request, jsonify

app = Flask(__name__)


@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint."""
    return jsonify({
        'status': 'healthy',
        'service': 'service-a'
    })


@app.route('/hello', methods=['GET'])
def hello():
    """Hello endpoint."""
    return jsonify({
        'message': 'Hello from Python Service A',
        'request_id': request.headers.get('X-Request-ID', 'unknown'),
        'user_id': request.headers.get('X-User-ID', 'unknown')
    })


@app.route('/echo', methods=['POST', 'GET', 'PUT', 'DELETE', 'PATCH'])
def echo():
    """Echo back request details."""
    # Get request body
    body = None
    if request.data:
        try:
            body = request.get_json(force=True)
        except Exception:
            body = request.data.decode('utf-8')
    
    return jsonify({
        'service': 'service-a',
        'method': request.method,
        'path': request.path,
        'query_params': dict(request.args),
        'headers': {k: v for k, v in request.headers if k.lower() not in ['host', 'content-length']},
        'body': body
    })


if __name__ == '__main__':
    port = int(os.environ.get('PORT', 6000))
    app.run(host='0.0.0.0', port=port, debug=False)
