package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"

	"github.com/ayitas/recension/api/internal/auth"
	"github.com/ayitas/recension/api/internal/blobstore"
)

func (s *Server) requireBlobs(w http.ResponseWriter) bool {
	if s.blobs == nil {
		writeErr(w, http.StatusServiceUnavailable, "blob storage not configured")
		return false
	}
	return true
}

func (s *Server) handlePutBlob(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.authenticateClient(r); !ok {
		writeErr(w, http.StatusUnauthorized, "invalid api key")
		return
	}
	if !s.requireBlobs(w) {
		return
	}
	digest := r.PathValue("digest")
	if !blobstore.ValidDigest(digest) {
		writeErr(w, http.StatusBadRequest, "digest must be sha256:<64 hex chars>")
		return
	}

	limit := s.config.MaxBlobBytes
	if limit <= 0 {
		limit = 64 << 20
	}
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, http.StatusRequestEntityTooLarge, "blob too large")
		return
	}
	sum := sha256.Sum256(data)
	got := blobstore.DigestPrefix + hex.EncodeToString(sum[:])
	if got != digest {
		writeErr(w, http.StatusBadRequest, "digest mismatch")
		return
	}

	mime := r.Header.Get("Content-Type")
	if mime == "" || mime == "application/octet-stream" {
		if ct := r.Header.Get("X-Recension-Mime"); ct != "" {
			mime = ct
		}
	}
	if err := s.blobs.Put(r.Context(), digest, data, mime); err != nil {
		writeErr(w, http.StatusInternalServerError, "could not store blob")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"digest": digest,
		"size":   len(data),
		"mime":   mime,
	})
}

func (s *Server) handleHeadBlob(w http.ResponseWriter, r *http.Request) {
	if !s.authorizeBlobRead(w, r) {
		return
	}
	if !s.requireBlobs(w) {
		return
	}
	digest := r.PathValue("digest")
	ok, err := s.blobs.Exists(r.Context(), digest)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid digest")
		return
	}
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleGetBlob(w http.ResponseWriter, r *http.Request) {
	if !s.authorizeBlobRead(w, r) {
		return
	}
	if !s.requireBlobs(w) {
		return
	}
	digest := r.PathValue("digest")
	rc, mime, size, err := s.blobs.Get(r.Context(), digest)
	if err != nil {
		writeErr(w, http.StatusNotFound, "blob not found")
		return
	}
	defer rc.Close()
	if mime == "" {
		mime = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	w.Header().Set("X-Recension-Digest", digest)
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, rc)
}

func (s *Server) authorizeBlobRead(w http.ResponseWriter, r *http.Request) bool {
	if _, ok := s.authenticateClient(r); ok {
		return true
	}
	tok, err := auth.BearerToken(r.Header.Get("Authorization"))
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "authentication required")
		return false
	}
	claims, err := s.tokens.Parse(tok)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "authentication required")
		return false
	}
	if _, ok := s.store.UserByID(claims.UserID); !ok {
		writeErr(w, http.StatusUnauthorized, "authentication required")
		return false
	}
	return true
}
