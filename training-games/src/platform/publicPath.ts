const publicBasePath = (() => {
  const path = new URL('.', window.location.href).pathname;
  return path.endsWith('/') ? path.slice(0, -1) : path;
})();

export function publicAssetPath(path: string): string {
  const normalizedPath = path.startsWith('/') ? path.slice(1) : path;
  return `${publicBasePath}/${normalizedPath}`;
}
