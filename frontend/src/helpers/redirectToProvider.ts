import {getFullBaseUrl} from '@/helpers/getFullBaseUrl'
import {createRandomID} from '@/helpers/randomId'
import type {IProvider} from '@/types/IProvider'
import {parseURL} from 'ufo'

export function getRedirectUrlFromCurrentFrontendPath(provider: IProvider): string {
	// We're not using the redirect url provided by the server to allow redirects when using the electron app.
	// The implications are not quite clear yet hence the logic to pass in another redirect url still exists.
	const url = parseURL(window.location.href)
	const base = getFullBaseUrl()
	return `${url.protocol}//${url.host}${base}auth/openid/${provider.key}`
}

export const OIDC_AUTH_SUCCESS_MESSAGE = 'vikunja-oidc-auth-success'

// Returns a popup window when the auth was opened in a popup (i.e. we're running in an iframe),
// or null when we navigated directly (normal top-level flow).
export const redirectToProvider = (provider: IProvider): Window | null => {

	const redirectUrl = getRedirectUrlFromCurrentFrontendPath(provider)
	const state = createRandomID(24)
	localStorage.setItem('state', state)

	let scope = 'openid email profile'
	if (provider.scope !== null){
		scope = provider.scope
	}
	const authUrl = `${provider.authUrl}?client_id=${provider.clientId}&redirect_uri=${redirectUrl}&response_type=code&scope=${scope}&state=${state}`

	// When embedded in an iframe, providers like GitHub set X-Frame-Options / CSP that block
	// the redirect inside the frame. Open a popup instead so the OAuth flow runs top-level.
	if (window.self !== window.top) {
		return window.open(authUrl, 'vikunja-oauth', 'width=600,height=700')
	}

	window.location.href = authUrl
	return null
}

export const redirectToProviderOnLogout = (provider: IProvider) => {
	if (provider.logoutUrl.length > 0) {
		window.location.href = `${provider.logoutUrl}`
	}
}
