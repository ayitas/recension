#!/usr/bin/env python3
"""Recolor Recension brand teal → ocean blue in PNG assets (vectorized)."""

from __future__ import annotations

import shutil
from pathlib import Path

import numpy as np
from PIL import Image

ROOT = Path(__file__).resolve().parents[1]
BRAND = ROOT / "docs" / "brand"
STATIC = ROOT / "web" / "static"
SIZES = BRAND / "sizes"

# Ocean accent hue (~199°)
OCEAN_H = 199.0 / 360.0


def rgb_to_hsv_np(rgb: np.ndarray) -> tuple[np.ndarray, np.ndarray, np.ndarray]:
	r, g, b = rgb[..., 0], rgb[..., 1], rgb[..., 2]
	maxc = np.maximum(np.maximum(r, g), b)
	minc = np.minimum(np.minimum(r, g), b)
	v = maxc
	del_ = maxc - minc
	s = np.where(maxc == 0, 0.0, del_ / np.maximum(maxc, 1e-12))
	rc = (maxc - r) / np.maximum(del_, 1e-12)
	gc = (maxc - g) / np.maximum(del_, 1e-12)
	bc = (maxc - b) / np.maximum(del_, 1e-12)
	h = np.zeros_like(maxc)
	h = np.where((maxc == r) & (del_ > 0), (bc - gc), h)
	h = np.where((maxc == g) & (del_ > 0), 2.0 + (rc - bc), h)
	h = np.where((maxc == b) & (del_ > 0), 4.0 + (gc - rc), h)
	h = (h / 6.0) % 1.0
	h = np.where(del_ == 0, 0.0, h)
	return h, s, v


def hsv_to_rgb_np(h: np.ndarray, s: np.ndarray, v: np.ndarray) -> np.ndarray:
	i = np.floor(h * 6.0).astype(np.int32)
	f = h * 6.0 - i
	p = v * (1.0 - s)
	q = v * (1.0 - s * f)
	t = v * (1.0 - s * (1.0 - f))
	i_mod = i % 6
	r = np.choose(i_mod, [v, q, p, p, t, v])
	g = np.choose(i_mod, [t, v, v, q, p, p])
	b = np.choose(i_mod, [p, p, t, v, v, q])
	return np.stack([r, g, b], axis=-1)


def recolor_rgba(arr: np.ndarray) -> tuple[np.ndarray, int]:
	rgb = arr[..., :3].astype(np.float64) / 255.0
	a = arr[..., 3]
	h, s, v = rgb_to_hsv_np(rgb)
	deg = h * 360.0
	r8, g8, b8 = arr[..., 0], arr[..., 1], arr[..., 2]

	# Brand seal teal: hue ~160–185, chromatic, not parchment
	# Also catch dark teal (#0a4d4b family) and light seafoam accents
	mask = (
		(a >= 8)
		& (s >= 0.15)
		& (v >= 0.06)
		& (v <= 0.85)
		& (deg >= 150.0)
		& (deg <= 190.0)
		& (g8 + 5 >= r8)  # green channel dominates red (teal)
	)

	changed = int(mask.sum())
	if changed == 0:
		return arr, 0

	# Slightly boost sat toward ocean stamp; keep value (shading)
	s2 = np.clip(np.maximum(s, 0.45) * 1.08, 0, 1)
	# Lift very dark seals a touch toward accent readability
	v2 = np.where(v < 0.22, np.minimum(v * 1.15, 0.28), v)
	h2 = np.full_like(h, OCEAN_H)
	new_rgb = hsv_to_rgb_np(h2, s2, v2)
	out = arr.copy()
	out[mask, :3] = np.clip(new_rgb[mask] * 255.0, 0, 255).astype(np.uint8)
	return out, changed


def recolor_file(src: Path, dst: Path | None = None) -> None:
	dst = dst or src
	im = Image.open(src).convert("RGBA")
	arr = np.array(im)
	out, changed = recolor_rgba(arr)
	Image.fromarray(out).save(dst)
	print(f"{src.relative_to(ROOT)} ({changed} px)", flush=True)


def resize(src: Path, dst: Path, size: int) -> None:
	im = Image.open(src).convert("RGBA")
	im = im.resize((size, size), Image.Resampling.LANCZOS)
	dst.parent.mkdir(parents=True, exist_ok=True)
	im.save(dst)
	print(f"resize {size} → {dst.relative_to(ROOT)}", flush=True)


def main() -> None:
	for name in [
		"recension-logo-mark.png",
		"recension-logo-mark-flat.png",
		"recension-logo-mark-dark.png",
		"recension-logo-mark-mono.png",
		"recension-logo-lockup.png",
		"recension-github-banner.png",
	]:
		path = BRAND / name
		if path.exists():
			recolor_file(path)

	flat = BRAND / "recension-logo-mark-flat.png"
	mark = BRAND / "recension-logo-mark.png"
	lockup = BRAND / "recension-logo-lockup.png"

	if flat.exists():
		resize(flat, STATIC / "favicon.png", 48)
		resize(flat, STATIC / "favicon-32.png", 32)
		resize(flat, STATIC / "favicon-16.png", 16)
		resize(flat, STATIC / "apple-touch-icon.png", 180)
		resize(flat, SIZES / "favicon-16.png", 16)
		resize(flat, SIZES / "favicon-32.png", 32)
		resize(flat, SIZES / "favicon-64.png", 64)
		resize(flat, SIZES / "apple-touch-icon.png", 180)
		resize(flat, SIZES / "github-avatar.png", 200)
		resize(flat, SIZES / "logo-256.png", 256)
		resize(flat, SIZES / "logo-512.png", 512)

	if mark.exists():
		im = Image.open(mark).convert("RGBA")
		im.thumbnail((256, 256), Image.Resampling.LANCZOS)
		im.save(STATIC / "logo-mark.png")
		print("web/static/logo-mark.png from detailed mark", flush=True)

	if lockup.exists():
		shutil.copy2(lockup, STATIC / "logo-lockup.png")
		print("copy → web/static/logo-lockup.png", flush=True)


if __name__ == "__main__":
	main()
