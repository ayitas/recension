const TOKEN_KEY = 'recension_token';

export type AuthUser = {
	id: string;
	email: string;
	apiKey: string;
};

export function getToken(): string | null {
	if (typeof localStorage === 'undefined') return null;
	return localStorage.getItem(TOKEN_KEY);
}

export function setSession(token: string) {
	localStorage.setItem(TOKEN_KEY, token);
}

export function clearSession() {
	localStorage.removeItem(TOKEN_KEY);
}

export function isLoggedIn(): boolean {
	return !!getToken();
}
