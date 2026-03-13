import hljs from 'highlight.js/lib/core';
import json from 'highlight.js/lib/languages/json';
import xml from 'highlight.js/lib/languages/xml';
import graphql from 'highlight.js/lib/languages/graphql';

hljs.registerLanguage('json', json);
hljs.registerLanguage('xml', xml);
hljs.registerLanguage('graphql', graphql);

export function getStatusColor(code: number): string {
  if (code >= 200 && code < 300) return '#49cc90';
  if (code >= 300 && code < 400) return '#fca130';
  if (code >= 400 && code < 500) return '#f93e3e';
  if (code >= 500) return '#f93e3e';
  return '#999';
}

export function formatHeaders(headers: Record<string, string> | null | undefined): string {
  if (!headers) return '';
  return Object.entries(headers)
    .map(([k, v]) => `${k}: ${v}`)
    .join('\n');
}

export function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(2)}s`;
}

export function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
}

export function getResponseSize(body: string | undefined): number {
  if (!body) return 0;
  return new Blob([body]).size;
}

export function getHighlightedBody(body: string | undefined, contentType: string | undefined): string {
  if (!body) return '';
  const type = contentType?.toLowerCase() || '';
  try {
    if (type.includes('json')) return hljs.highlight(body, { language: 'json' }).value;
    if (type.includes('graphql')) return hljs.highlight(body, { language: 'graphql' }).value;
    if (type.includes('xml') || type.includes('html')) return hljs.highlight(body, { language: 'xml' }).value;
  } catch (e) {
    console.warn('Highlight error:', e);
  }
  return body
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}

// Inject <mark> tags into hljs HTML without touching tag internals.
// Splits on HTML tags/entities so the regex only runs on plain text content.
export function applySearchHighlights(html: string, term: string): { html: string; count: number } {
  if (!term.trim()) return { html, count: 0 };
  const escaped = term.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const re = new RegExp(escaped, 'gi');
  let count = 0;
  // Even indices = text content, odd indices = tags / &entities;
  const parts = html.split(/(<[^>]*>|&[^;]+;)/);
  const out = parts.map((part, i) => {
    if (i % 2 !== 0) return part;
    return part.replace(re, match => {
      const idx = count++;
      return `<mark class="search-mark" data-idx="${idx}">${match}</mark>`;
    });
  });
  return { html: out.join(''), count };
}
