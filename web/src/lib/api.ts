import { getToken, clearSession, setSession, type AuthUser } from '$lib/auth';

const API_BASE = '';

async function api<T>(path: string, init?: RequestInit): Promise<T> {
	const headers: Record<string, string> = {
		Accept: 'application/json',
		...(init?.headers as Record<string, string> | undefined)
	};
	const token = getToken();
	if (token) {
		headers.Authorization = `Bearer ${token}`;
	}
	const res = await fetch(`${API_BASE}${path}`, {
		...init,
		headers
	});
	if (res.status === 401 && !path.startsWith('/v1/auth/login') && !path.startsWith('/v1/auth/signup')) {
		clearSession();
	}
	if (!res.ok) {
		const text = await res.text();
		throw new Error(formatApiError(text, res.statusText));
	}
	if (res.status === 204) {
		return undefined as T;
	}
	return res.json() as Promise<T>;
}

export type Team = { id: string; slug: string; name: string };
export type TeamRole = 'owner' | 'admin' | 'member' | 'viewer';
export type TeamMember = { userId: string; email: string; role: TeamRole };

export type Suite = {
	id: string;
	teamId: string;
	slug: string;
	name: string;
	baselineBatchId: string;
};
export type Batch = {
	id: string;
	suiteId: string;
	slug: string;
	sealedAt?: string | null;
	submittedAt: string;
	meta: Record<string, unknown>;
	testcaseCount?: number;
	passCount?: number;
	diffCount?: number;
	sentCount?: number;
	avgScore?: number;
	isBaseline?: boolean;
};

export type Cell = {
	name: string;
	score?: number;
	srcType?: string;
	dstType?: string;
	srcValue?: string;
	dstValue?: string;
};

export type Cellar = {
	commonKeys: Cell[];
	newKeys: Cell[];
	missingKeys: Cell[];
};

export type Comparison = {
	overview: {
		keysCountCommon: number;
		keysCountFresh: number;
		keysCountMissing: number;
		keysScore: number;
		metricsCountCommon: number;
		metricsCountFresh: number;
		metricsCountMissing: number;
		metricsDurationCommonSrc: number;
		metricsDurationCommonDst: number;
	};
	src: { team: string; suite: string; version: string; testcase: string; builtAt: string };
	dst: { team: string; suite: string; version: string; testcase: string; builtAt: string };
	results: Cellar;
	asserts: Cellar;
	metrics: Cellar;
};

export type BatchDetail = {
	batch: Batch;
	baselineBatchId: string;
	elements: Array<{
		testcase: string;
		builtAt: string;
		verdict: string;
		score: number;
		comparison?: Comparison;
	}>;
};

export type ElementDetail = {
	testcase: string;
	verdict: string;
	score: number;
	batch: Batch;
	baselineBatchId: string;
	message: {
		metadata: { team: string; suite: string; version: string; testcase: string; builtAt: string };
		results: Array<{ key: string; kind: string; value: unknown }>;
		metrics: Array<{ key: string; value: number }>;
	};
	comparison: Comparison;
};

type SessionResponse = {
	token: string;
	expiresAt: string;
	user: AuthUser;
};

export const login = async (email: string, password: string) => {
	const data = await api<SessionResponse>('/v1/auth/login', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ email, password })
	});
	setSession(data.token);
	return data.user;
};

export const signup = async (email: string, password: string) => {
	const data = await api<SessionResponse>('/v1/auth/signup', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ email, password })
	});
	setSession(data.token);
	return data.user;
};

export const me = () => api<AuthUser>('/v1/auth/me');
export const rotateAPIKey = () =>
	api<AuthUser>('/v1/auth/api-key/rotate', { method: 'POST' });

export const listTeams = () => api<Team[]>('/v1/teams');
export const createTeam = (name: string, slug: string) =>
	api<Team>('/v1/teams', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ name, slug })
	});
export const listSuites = (team: string) => api<Suite[]>(`/v1/teams/${team}/suites`);
export const createSuite = (team: string, name: string, slug: string) =>
	api<Suite>(`/v1/teams/${team}/suites`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ name, slug })
	});

export const listMembers = (team: string) => api<TeamMember[]>(`/v1/teams/${team}/members`);
export const addMember = (team: string, email: string, role: TeamRole) =>
	api<TeamMember>(`/v1/teams/${team}/members`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ email, role })
	});
export const updateMemberRole = (team: string, userId: string, role: TeamRole) =>
	api<TeamMember>(`/v1/teams/${team}/members/${userId}`, {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ role })
	});
export const removeMember = (team: string, userId: string) =>
	api<void>(`/v1/teams/${team}/members/${userId}`, { method: 'DELETE' });

export const ROLE_RANK: Record<TeamRole, number> = {
	viewer: 1,
	member: 2,
	admin: 3,
	owner: 4
};

export function roleAtLeast(have: TeamRole | undefined, need: TeamRole): boolean {
	return !!have && ROLE_RANK[have] >= ROLE_RANK[need];
}
export const listBatches = (team: string, suite: string) =>
	api<Batch[]>(`/v1/teams/${team}/suites/${suite}/batches`);
export const getBatch = (team: string, suite: string, batch: string) =>
	api<BatchDetail>(`/v1/teams/${team}/suites/${suite}/batches/${batch}`);
export const getElement = (team: string, suite: string, batch: string, element: string) =>
	api<ElementDetail>(
		`/v1/teams/${team}/suites/${suite}/batches/${batch}/elements/${encodeURIComponent(element)}`
	);
export const promoteBatch = (team: string, suite: string, batch: string) =>
	api<void>(`/v1/batch/${team}/${suite}/${batch}/promote`, { method: 'POST' });

export function shortDigest(digest: string): string {
	const hex = digest.startsWith('sha256:') ? digest.slice('sha256:'.length) : digest;
	if (hex.length <= 16) return digest;
	return `sha256:${hex.slice(0, 8)}…${hex.slice(-4)}`;
}

export async function downloadBlob(digest: string, filename?: string): Promise<void> {
	const token = getToken();
	const headers: Record<string, string> = { Accept: '*/*' };
	if (token) {
		headers.Authorization = `Bearer ${token}`;
	}
	const res = await fetch(`/v1/blobs/${encodeURIComponent(digest)}`, { headers });
	if (res.status === 401) {
		clearSession();
		throw new Error('authentication required');
	}
	if (!res.ok) {
		throw new Error((await res.text()) || res.statusText);
	}
	const blob = await res.blob();
	const url = URL.createObjectURL(blob);
	const a = document.createElement('a');
	a.href = url;
	a.download = filename || `${shortDigest(digest).replace(/[^\w.-]+/g, '_')}.bin`;
	document.body.appendChild(a);
	a.click();
	a.remove();
	URL.revokeObjectURL(url);
}

export function slugify(input: string): string {
	return input
		.trim()
		.toLowerCase()
		.replace(/[_\s]+/g, '-')
		.replace(/[^a-z0-9-]/g, '')
		.replace(/-+/g, '-')
		.replace(/^-|-$/g, '');
}

function formatApiError(text: string, fallback: string): string {
	if (!text) return fallback;
	try {
		const parsed = JSON.parse(text) as { errors?: string[]; error?: string };
		if (Array.isArray(parsed.errors) && parsed.errors.length > 0) {
			return parsed.errors.join('; ');
		}
		if (parsed.error) return parsed.error;
	} catch {
		/* raw text */
	}
	return text;
}