export function formatValue(value: any): string {
  if (typeof value === 'object') {
    return JSON.stringify(value, null, 2);
  }
  const str = String(value);
  if (str.length > 40) {
    return str.substring(0, 15) + '...' + str.substring(str.length - 10);
  }
  return str;
}
