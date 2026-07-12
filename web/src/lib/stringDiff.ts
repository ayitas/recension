export type DiffSeg = { text: string; kind: 'eq' | 'ins' | 'del' };

/** Character-level LCS diff for side-by-side string highlighting. */
export function charDiff(a: string, b: string): { src: DiffSeg[]; dst: DiffSeg[] } {
	if (a === b) {
		return {
			src: a ? [{ text: a, kind: 'eq' }] : [],
			dst: b ? [{ text: b, kind: 'eq' }] : []
		};
	}
	// Avoid O(n²) blowups on huge payloads.
	if (a.length * b.length > 400_000) {
		return {
			src: a ? [{ text: a, kind: 'del' }] : [],
			dst: b ? [{ text: b, kind: 'ins' }] : []
		};
	}

	const n = a.length;
	const m = b.length;
	const dp: Uint16Array[] = Array.from({ length: n + 1 }, () => new Uint16Array(m + 1));
	for (let i = 1; i <= n; i++) {
		for (let j = 1; j <= m; j++) {
			if (a[i - 1] === b[j - 1]) {
				dp[i][j] = dp[i - 1][j - 1] + 1;
			} else {
				dp[i][j] = Math.max(dp[i - 1][j], dp[i][j - 1]);
			}
		}
	}

	const src: DiffSeg[] = [];
	const dst: DiffSeg[] = [];
	let i = n;
	let j = m;
	const push = (side: DiffSeg[], text: string, kind: DiffSeg['kind']) => {
		if (!text) return;
		const last = side[side.length - 1];
		if (last && last.kind === kind) {
			last.text = text + last.text;
		} else {
			side.push({ text, kind });
		}
	};

	while (i > 0 || j > 0) {
		if (i > 0 && j > 0 && a[i - 1] === b[j - 1]) {
			push(src, a[i - 1], 'eq');
			push(dst, b[j - 1], 'eq');
			i--;
			j--;
		} else if (j > 0 && (i === 0 || dp[i][j - 1] >= dp[i - 1][j])) {
			push(dst, b[j - 1], 'ins');
			j--;
		} else {
			push(src, a[i - 1], 'del');
			i--;
		}
	}

	src.reverse();
	dst.reverse();
	return { src, dst };
}

export function shouldHighlightString(srcType?: string, dstType?: string): boolean {
	return srcType === 'string' || dstType === 'string';
}
