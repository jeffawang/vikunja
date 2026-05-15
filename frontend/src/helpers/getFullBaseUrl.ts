/**
 * Get full base URL
 * - including path
 * - will always end with a trailing slash
 *
 * Uses window.VIKUNJA_BASE_PATH when injected by the server (runtime),
 * falling back to the build-time __VIKUNJA_BASE_PATH__ constant.
 */
export function getFullBaseUrl() {
	const base = (globalThis as { VIKUNJA_BASE_PATH?: string }).VIKUNJA_BASE_PATH || __VIKUNJA_BASE_PATH__
	return base.endsWith('/') ? base : base + '/'
}
