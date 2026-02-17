package httpapi

import (
	"fmt"
	"io"
	"net/http"

	"ticketing-system/backend/internal/store"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h *API) ListTicketAttachments(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, ticketId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	ticketUUID := uuid.UUID(ticketId)
	attachments, err := h.store.ListAttachments(r.Context(), ticketUUID)
	if handleListError(w, r, err, "attachments", "attachment_list") {
		return
	}
	writeJSON(w, http.StatusOK, attachmentListResponse{Items: mapSlice(attachments, mapAttachment)})
}

func (h *API) UploadTicketAttachment(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, ticketId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectUUID, roleContributor) {
		return
	}
	if h.blob == nil {
		writeError(w, http.StatusServiceUnavailable, "blob_unavailable", "file storage not configured")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadSize)
	if err := r.ParseMultipartForm(h.maxUploadSize); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "file_too_large", "file exceeds maximum upload size")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing_file", "file field is required")
		return
	}
	defer file.Close()

	userID, err := h.currentUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "user not found")
		return
	}
	user, ok := authUser(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "user not found")
		return
	}

	ticketUUID := uuid.UUID(ticketId)
	fileID := uuid.New()
	storageKey := fmt.Sprintf("attachments/%s/%s/%s", ticketUUID, fileID, header.Filename)
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if err := h.blob.Put(r.Context(), storageKey, file, header.Size, contentType); err != nil {
		logRequestError(r, "attachment_upload_failed", err)
		writeError(w, http.StatusInternalServerError, "upload_failed", "failed to store file")
		return
	}

	att, err := h.store.CreateAttachment(r.Context(), ticketUUID, store.AttachmentCreateInput{
		Filename:       header.Filename,
		ContentType:    contentType,
		Size:           header.Size,
		StorageKey:     storageKey,
		UploadedBy:     userID,
		UploadedByName: user.Name,
	})
	if handleDBErrorWithCode(w, r, err, "attachment", "attachment_create", "attachment_create_failed") {
		return
	}

	writeJSON(w, http.StatusCreated, mapAttachment(att))
}

func (h *API) DownloadTicketAttachment(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, ticketId openapi_types.UUID, attachmentId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	if h.blob == nil {
		writeError(w, http.StatusServiceUnavailable, "blob_unavailable", "file storage not configured")
		return
	}

	attUUID := uuid.UUID(attachmentId)
	att, err := h.store.GetAttachment(r.Context(), attUUID)
	if handleDBError(w, r, err, "attachment", "attachment_load") {
		return
	}
	ticketUUID := uuid.UUID(ticketId)
	if att.TicketID != ticketUUID {
		writeError(w, http.StatusNotFound, "not_found", "attachment not found")
		return
	}

	reader, err := h.blob.Get(r.Context(), att.StorageKey)
	if err != nil {
		logRequestError(r, "attachment_download_failed", err)
		writeError(w, http.StatusInternalServerError, "download_failed", "failed to retrieve file")
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", att.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, att.Filename))
	io.Copy(w, reader)
}

func (h *API) DeleteTicketAttachment(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, ticketId openapi_types.UUID, attachmentId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectUUID, roleContributor) {
		return
	}

	attUUID := uuid.UUID(attachmentId)
	att, err := h.store.GetAttachment(r.Context(), attUUID)
	if handleDBError(w, r, err, "attachment", "attachment_load") {
		return
	}
	ticketUUID := uuid.UUID(ticketId)
	if att.TicketID != ticketUUID {
		writeError(w, http.StatusNotFound, "not_found", "attachment not found")
		return
	}

	if err := h.store.DeleteAttachment(r.Context(), attUUID); handleDeleteError(w, r, err, "attachment", "attachment_delete") {
		return
	}

	if h.blob != nil {
		_ = h.blob.Delete(r.Context(), att.StorageKey)
	}

	w.WriteHeader(http.StatusNoContent)
}
