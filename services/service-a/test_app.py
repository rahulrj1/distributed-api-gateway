"""
Unit tests for Service A

Run locally (from services/service-a/):
    pytest test_app.py -v

Run via Docker (from project root):
    docker build -t service-a ./services/service-a
    docker run --rm service-a pytest test_app.py -v
"""

import pytest
from app import app


@pytest.fixture
def client():
    """Create test client."""
    app.config['TESTING'] = True
    with app.test_client() as client:
        yield client


class TestHealth:
    """Tests for /health endpoint."""
    
    def test_health_returns_200(self, client):
        response = client.get('/health')
        assert response.status_code == 200
    
    def test_health_returns_json(self, client):
        response = client.get('/health')
        data = response.get_json()
        assert data['status'] == 'healthy'
        assert data['service'] == 'service-a'


class TestHello:
    """Tests for /hello endpoint."""
    
    def test_hello_returns_200(self, client):
        response = client.get('/hello')
        assert response.status_code == 200
    
    def test_hello_returns_message(self, client):
        response = client.get('/hello')
        data = response.get_json()
        assert 'message' in data
        assert 'Python Service A' in data['message']


class TestEcho:
    """Tests for /echo endpoint."""
    
    def test_echo_post_returns_200(self, client):
        response = client.post('/echo', json={'test': 'data'})
        assert response.status_code == 200
    
    def test_echo_returns_method(self, client):
        response = client.post('/echo', json={'test': 'data'})
        data = response.get_json()
        assert data['method'] == 'POST'
    
    def test_echo_returns_body(self, client):
        response = client.post('/echo', json={'key': 'value'})
        data = response.get_json()
        assert data['body'] == {'key': 'value'}
    
    def test_echo_get_works(self, client):
        response = client.get('/echo')
        assert response.status_code == 200
        assert response.get_json()['method'] == 'GET'
