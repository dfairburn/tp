import type { app } from '../../wailsjs/go/models';

export function makeErrorResponse(err: unknown): app.HTTPResponse {
  return {
    statusCode: 0,
    status: 'Error',
    headers: {},
    body: '',
    contentType: '',
    duration: 0,
    error: String(err),
  } as app.HTTPResponse;
}

export function getMethodColor(method: string): string {
  const colors: Record<string, string> = {
    GET: '#61affe',
    POST: '#49cc90',
    PUT: '#fca130',
    PATCH: '#50e3c2',
    DELETE: '#f93e3e',
    HEAD: '#9012fe',
    OPTIONS: '#0d5aa7',
  };
  return colors[method?.toUpperCase()] || '#999';
}
