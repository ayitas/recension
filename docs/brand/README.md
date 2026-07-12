# Brand assets

Hybrid web UI tokens (Recension ocean accent on parchment):

| Token | Hex (light) | Dark |
|-------|-------------|------|
| Canvas / parchment | `#f3efe6` | `#171512` |
| Bone (inset) | `#ebe4d7` | `#100e0c` |
| Card | `#fffdf8` | `#211e19` |
| Ink | `#1c1915` | `#f3efe6` |
| Body | `#3a342c` | `#d4cdc2` |
| Muted | `#6b6358` | `#a39a8c` |
| Hairline | `rgba(28,25,21,0.12)` | `rgba(243,239,230,0.12)` |
| Ocean accent | `#0b4f6c` | `#5ba4c4` |
| Accent deep | `#083d54` | `#7bb8d0` |
| Accent soft | `#d5e8f0` | `rgba(91,164,196,0.18)` |
| Surface dark (code wells) | `#202020` | `#2a2620` |

Typography: **Bricolage Grotesque** (display/brand), **Source Sans 3** (UI), **JetBrains Mono** (code). Interactive controls use full pill radius; content cards/tables use 10px.

Accent stamp is **ocean blue** (`#0b4f6c`); semantic pass/diff greens stay separate. To re-apply seal recolor on brand PNGs: `scripts/recolor-brand-ocean.py` (needs Pillow + NumPy).

## Files

| File | Use |
|------|-----|
| `recension-logo-mark.png` | Primary mark (detailed) |
| `recension-logo-mark-flat.png` | Favicon / small sizes |
| `recension-logo-mark-dark.png` | Dark backgrounds |
| `recension-logo-mark-mono.png` | Print / single-color |
| `recension-logo-lockup.png` | Wordmark lockup |
| `recension-github-banner.png` | GitHub social / OG |
| `sizes/` | Pre-resized favicon, apple-touch, avatar |

GitHub: set **avatar** to `sizes/github-avatar.png`, **social preview** to `recension-github-banner.png`.
