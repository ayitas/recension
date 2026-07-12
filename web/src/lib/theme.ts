const KEY = 'recension_theme';

export type Theme = 'light' | 'dark';

export function getStoredTheme(): Theme {
	if (typeof localStorage === 'undefined') return 'light';
	const v = localStorage.getItem(KEY);
	return v === 'dark' ? 'dark' : 'light';
}

export function applyTheme(theme: Theme) {
	if (typeof document === 'undefined') return;
	document.documentElement.dataset.theme = theme;
	localStorage.setItem(KEY, theme);
	const meta = document.querySelector('meta[name="theme-color"]');
	if (meta) {
		meta.setAttribute('content', theme === 'dark' ? '#171512' : '#0b4f6c');
	}
}

export function toggleTheme(): Theme {
	const next: Theme = getStoredTheme() === 'dark' ? 'light' : 'dark';
	applyTheme(next);
	return next;
}
