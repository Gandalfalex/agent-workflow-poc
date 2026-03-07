package httpapi

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"ticketing-system/backend/internal/store"
)

// --- Project custom field definitions ---

func (h *API) ListCustomFields(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	fields, err := h.store.ListCustomFields(r.Context(), projectUUID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "custom_fields_load_error", "Failed to load custom fields")
		return
	}
	writeJSON(w, http.StatusOK, customFieldDefinitionListResponse{Items: mapCustomFields(fields)})
}

func (h *API) CreateCustomField(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectUUID, roleAdmin) {
		return
	}
	req, ok := decodeJSON[customFieldDefinitionCreateRequest](w, r, "custom_field_create")
	if !ok {
		return
	}

	options := []store.CustomFieldOption{}
	if req.Options != nil {
		for _, o := range *req.Options {
			options = append(options, store.CustomFieldOption{Value: o.Value, Label: o.Label})
		}
	}
	schemas := parseSchemaInputs(req.Schemas)

	required := false
	if req.Required != nil {
		required = *req.Required
	}

	field, err := h.store.CreateCustomField(r.Context(), projectUUID, store.CustomFieldCreateInput{
		Name:      req.Name,
		Label:     req.Label,
		FieldType: string(req.FieldType),
		Options:   options,
		Required:  required,
		Schemas:   schemas,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "custom_field_create_error", "Failed to create custom field")
		return
	}
	writeJSON(w, http.StatusCreated, mapCustomField(field))
}

func (h *API) UpdateCustomField(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, fieldId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectUUID, roleAdmin) {
		return
	}
	req, ok := decodeJSON[customFieldDefinitionUpdateRequest](w, r, "custom_field_update")
	if !ok {
		return
	}

	// Fetch existing to preserve immutable fields
	existing, err := h.store.GetCustomField(r.Context(), uuid.UUID(fieldId), projectUUID)
	if err != nil {
		if errors.Is(err, store.ErrCustomFieldNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Custom field not found")
		} else {
			writeError(w, http.StatusInternalServerError, "custom_field_load_error", "Failed to load custom field")
		}
		return
	}

	label := existing.Label
	if req.Label != nil {
		label = *req.Label
	}

	options := existing.Options
	if req.Options != nil {
		options = []store.CustomFieldOption{}
		for _, o := range *req.Options {
			options = append(options, store.CustomFieldOption{Value: o.Value, Label: o.Label})
		}
	}

	required := existing.Required
	if req.Required != nil {
		required = *req.Required
	}

	schemas := existing.Schemas
	if req.Schemas != nil {
		schemas = parseSchemaInputs(req.Schemas)
	}

	updated, err := h.store.UpdateCustomField(r.Context(), uuid.UUID(fieldId), projectUUID, store.CustomFieldUpdateInput{
		Label:    label,
		Options:  options,
		Required: required,
		Schemas:  schemas,
	})
	if err != nil {
		if errors.Is(err, store.ErrCustomFieldNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Custom field not found")
		} else {
			writeError(w, http.StatusInternalServerError, "custom_field_update_error", "Failed to update custom field")
		}
		return
	}
	writeJSON(w, http.StatusOK, mapCustomField(updated))
}

func (h *API) DeleteCustomField(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, fieldId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectUUID, roleAdmin) {
		return
	}
	if err := h.store.DeleteCustomField(r.Context(), uuid.UUID(fieldId), projectUUID); err != nil {
		if errors.Is(err, store.ErrCustomFieldNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Custom field not found")
		} else {
			writeError(w, http.StatusInternalServerError, "custom_field_delete_error", "Failed to delete custom field")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Ticket custom field values ---

func (h *API) GetTicketCustomFieldValues(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketUUID := uuid.UUID(id)
	// Verify ticket exists and get project for access check
	ticket, err := h.store.GetTicket(r.Context(), ticketUUID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "Ticket not found")
		return
	}
	if !h.requireProjectAccess(w, r, ticket.ProjectID) {
		return
	}
	values, err := h.store.GetCustomFieldValues(r.Context(), ticketUUID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "custom_field_values_load_error", "Failed to load custom field values")
		return
	}
	writeJSON(w, http.StatusOK, customFieldValueListResponse{Items: mapCustomFieldValues(values)})
}

func (h *API) SetTicketCustomFieldValues(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketUUID := uuid.UUID(id)
	ticket, err := h.store.GetTicket(r.Context(), ticketUUID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "Ticket not found")
		return
	}
	if !h.requireProjectRole(w, r, ticket.ProjectID, roleContributor) {
		return
	}

	req, ok := decodeJSON[customFieldValueUpsertRequest](w, r, "custom_field_values_set")
	if !ok {
		return
	}

	// Validate required fields for current state
	if ticket.StateID != uuid.Nil {
		fields, err := h.store.ListCustomFields(r.Context(), ticket.ProjectID)
		if err == nil && len(fields) > 0 {
			// Build provisional values map including incoming values
			incomingValues := make([]store.CustomFieldValue, 0, len(req.Values))
			for _, v := range req.Values {
				fid := uuid.UUID(v.FieldId)
				val := store.CustomFieldValue{FieldID: fid}
				if v.ValueText != nil {
					val.ValueText = v.ValueText
				}
				if v.ValueNumber != nil {
					f := float64(*v.ValueNumber)
					val.ValueNumber = &f
				}
				if v.ValueBoolean != nil {
					val.ValueBoolean = v.ValueBoolean
				}
				if v.ValueUserId != nil {
					uid := uuid.UUID(*v.ValueUserId)
					val.ValueUserID = &uid
				}
				incomingValues = append(incomingValues, val)
			}
			missing := store.CheckRequiredCustomFields(fields, incomingValues, ticket.Type, ticket.StateID)
			if len(missing) > 0 {
				msg := "Required custom fields are missing: "
				for i, m := range missing {
					if i > 0 {
						msg += ", "
					}
					msg += m
				}
				writeError(w, http.StatusUnprocessableEntity, "required_fields_missing", msg)
				return
			}
		}
	}

	inputs := make([]store.CustomFieldValueInput, 0, len(req.Values))
	for _, v := range req.Values {
		inp := store.CustomFieldValueInput{
			FieldID: uuid.UUID(v.FieldId),
		}
		if v.ValueText != nil {
			inp.ValueText = v.ValueText
		}
		if v.ValueNumber != nil {
			f := float64(*v.ValueNumber)
			inp.ValueNumber = &f
		}
		if v.ValueDate != nil {
			t := v.ValueDate.Time
			inp.ValueDate = &t
		}
		if v.ValueBoolean != nil {
			inp.ValueBoolean = v.ValueBoolean
		}
		if v.ValueUserId != nil {
			uid := uuid.UUID(*v.ValueUserId)
			inp.ValueUserID = &uid
		}
		inputs = append(inputs, inp)
	}

	updated, err := h.store.SetCustomFieldValues(r.Context(), ticketUUID, inputs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "custom_field_values_set_error", "Failed to set custom field values")
		return
	}
	writeJSON(w, http.StatusOK, customFieldValueListResponse{Items: mapCustomFieldValues(updated)})
}

// --- mapping helpers ---

func mapCustomField(f store.CustomFieldDefinition) customFieldDefinitionResponse {
	opts := make([]CustomFieldOption, 0, len(f.Options))
	for _, o := range f.Options {
		opts = append(opts, CustomFieldOption{Value: o.Value, Label: o.Label})
	}
	schemas := make([]CustomFieldTypeSchema, 0, len(f.Schemas))
	for _, sc := range f.Schemas {
		ids := make([]openapi_types.UUID, 0, len(sc.RequiredStateIDs))
		for _, id := range sc.RequiredStateIDs {
			ids = append(ids, toOpenapiUUID(id))
		}
		schemas = append(schemas, CustomFieldTypeSchema{
			TicketType:       TicketType(sc.TicketType),
			RequiredStateIds: ids,
		})
	}
	return customFieldDefinitionResponse{
		Id:        toOpenapiUUID(f.ID),
		ProjectId: toOpenapiUUID(f.ProjectID),
		Name:      f.Name,
		Label:     f.Label,
		FieldType: CustomFieldType(f.FieldType),
		Options:   opts,
		Order:     f.Order,
		Required:  f.Required,
		Schemas:   schemas,
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
	}
}

func mapCustomFields(fields []store.CustomFieldDefinition) []customFieldDefinitionResponse {
	out := make([]customFieldDefinitionResponse, 0, len(fields))
	for _, f := range fields {
		out = append(out, mapCustomField(f))
	}
	return out
}

func mapCustomFieldValue(v store.CustomFieldValue) customFieldValueResponse {
	r := customFieldValueResponse{
		FieldId:    toOpenapiUUID(v.FieldID),
		FieldName:  v.FieldName,
		FieldLabel: v.FieldLabel,
		FieldType:  CustomFieldType(v.FieldType),
	}
	r.ValueText = v.ValueText
	r.ValueBoolean = v.ValueBoolean
	if v.ValueNumber != nil {
		f := float32(*v.ValueNumber)
		r.ValueNumber = &f
	}
	if v.ValueDate != nil {
		d := openapi_types.Date{Time: *v.ValueDate}
		r.ValueDate = &d
	}
	if v.ValueUserID != nil {
		uid := toOpenapiUUID(*v.ValueUserID)
		r.ValueUserId = &uid
	}
	return r
}

func mapCustomFieldValues(values []store.CustomFieldValue) []customFieldValueResponse {
	out := make([]customFieldValueResponse, 0, len(values))
	for _, v := range values {
		out = append(out, mapCustomFieldValue(v))
	}
	return out
}

func parseSchemaInputs(schemas *[]CustomFieldTypeSchema) []store.CustomFieldSchema {
	if schemas == nil {
		return []store.CustomFieldSchema{}
	}
	out := make([]store.CustomFieldSchema, 0, len(*schemas))
	for _, sc := range *schemas {
		ids := make([]uuid.UUID, 0, len(sc.RequiredStateIds))
		for _, id := range sc.RequiredStateIds {
			ids = append(ids, uuid.UUID(id))
		}
		out = append(out, store.CustomFieldSchema{
			TicketType:       string(sc.TicketType),
			RequiredStateIDs: ids,
		})
	}
	return out
}
