import type { app } from '../../../wailsjs/go/models';

export const HTTP_METHODS = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'];
export const BODYLESS_METHODS = new Set(['GET', 'HEAD', 'OPTIONS']);

// Extract variable name from a single {{...}} block inner content,
// matching the same patterns as handlers/use-help.go ParseUsages.
export function extractVarFromBlock(inner: string): string | null {
  let m: RegExpMatchArray | null;
  if ((m = inner.match(/^\s*\.(\S+)\s*$/)))                           return m[1]; // {{.name}}
  if ((m = inner.match(/^\s*optional.*\.(\S+)\s*$/)))                 return m[1]; // {{optional ... .name}}
  if ((m = inner.match(/^\s*timestamp\s+\.(\S+)\s*$/)))               return m[1]; // {{timestamp .name}}
  if ((m = inner.match(/^\s*default\s+\.(\S+)\s+"[^"]*"\s*$/)))       return m[1]; // {{default .name "val"}}
  if ((m = inner.match(/^\s*default\s+"[^"]*"\s+\.(\S+)\s*$/)))       return m[1]; // {{default "val" .name}}
  return null;
}

export function extractVarsFromStr(str: string, vars: Set<string>) {
  const blockRegex = /\{\{([^}]+)\}\}/g;
  let block: RegExpExecArray | null;
  while ((block = blockRegex.exec(str)) !== null) {
    const v = extractVarFromBlock(block[1]);
    if (v) vars.add(v);
  }
}

export function extractVariables(tmpl: app.TemplateItem | null): string[] {
  if (!tmpl) return [];
  const vars = new Set<string>();
  if (tmpl.url) extractVarsFromStr(tmpl.url, vars);
  if (tmpl.headers) {
    for (const value of Object.values(tmpl.headers)) extractVarsFromStr(String(value), vars);
  }
  if (tmpl.body) extractVarsFromStr(tmpl.body, vars);
  return Array.from(vars).sort();
}

export function extractHeaderVars(tmpl: app.TemplateItem | null): string[] {
  if (!tmpl?.headers) return [];
  const vars = new Set<string>();
  for (const value of Object.values(tmpl.headers)) extractVarsFromStr(String(value), vars);
  return Array.from(vars).sort();
}

export function extractUrlOnlyVars(tmpl: app.TemplateItem | null): string[] {
  if (!tmpl?.url) return [];
  const vars = new Set<string>();
  extractVarsFromStr(tmpl.url, vars);
  return Array.from(vars).sort();
}

export function extractVarsFromStrings(url: string, headers: { key: string; value: string }[], body: string): string[] {
  const vars = new Set<string>();
  extractVarsFromStr(url, vars);
  for (const h of headers) extractVarsFromStr(h.value, vars);
  extractVarsFromStr(body, vars);
  return Array.from(vars).sort();
}

export function extractBodyVarsOnly(tmpl: app.TemplateItem | null): string[] {
  if (!tmpl?.body) return [];
  const vars = new Set<string>();
  extractVarsFromStr(tmpl.body, vars);
  return Array.from(vars).sort();
}
