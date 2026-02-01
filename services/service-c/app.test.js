/**
 * Unit tests for Service C
 * 
 * Run locally (from services/service-c/):
 *     npm test
 * 
 * Run via Docker (from project root):
 *     docker build -t service-c ./services/service-c
 *     docker run --rm service-c npm test
 */

const request = require('supertest');
const app = require('./app');

describe('Health endpoint', () => {
  test('GET /health returns 200', async () => {
    const res = await request(app).get('/health');
    expect(res.statusCode).toBe(200);
  });

  test('GET /health returns correct json', async () => {
    const res = await request(app).get('/health');
    expect(res.body.status).toBe('healthy');
    expect(res.body.service).toBe('service-c');
  });
});

describe('Hello endpoint', () => {
  test('GET /hello returns 200', async () => {
    const res = await request(app).get('/hello');
    expect(res.statusCode).toBe(200);
  });

  test('GET /hello returns message', async () => {
    const res = await request(app).get('/hello');
    expect(res.body.message).toContain('Node.js Service C');
  });
});

describe('Echo endpoint', () => {
  test('POST /echo returns 200', async () => {
    const res = await request(app)
      .post('/echo')
      .send({ test: 'data' });
    expect(res.statusCode).toBe(200);
  });

  test('POST /echo returns method', async () => {
    const res = await request(app)
      .post('/echo')
      .send({ test: 'data' });
    expect(res.body.method).toBe('POST');
  });

  test('POST /echo returns body', async () => {
    const res = await request(app)
      .post('/echo')
      .send({ key: 'value' });
    expect(res.body.body).toEqual({ key: 'value' });
  });

  test('GET /echo works', async () => {
    const res = await request(app).get('/echo');
    expect(res.statusCode).toBe(200);
    expect(res.body.method).toBe('GET');
  });
});
