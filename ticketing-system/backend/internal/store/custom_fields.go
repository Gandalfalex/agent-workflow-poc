package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ErrCustomFieldNotFound is returned when a custom field cannot be found.
var ErrCustomFieldNotFound = errors.New("custom field not found")

type CustomFieldOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type CustomFieldSchema struct {
	FieldID          uuid.UUID   `json:"fieldId"`
	TicketType       string      `json:"ticketType"`
	RequiredStateIDs []uuid.UUID `json:"requiredStateIds"`
}

type CustomFieldDefinition struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Name      string
	Label     string
	FieldType string
	Options   []CustomFieldOption
	Order     int
	Required  bool
	Schemas   []CustomFieldSchema
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CustomFieldCreateInput struct {
	Name      string
	Label     string
	FieldType string
	Options   []CustomFieldOption
	Required  bool
	Schemas   []CustomFieldSchema
}

type CustomFieldUpdateInput struct {
	Label   string
	Options []CustomFieldOption
	Required bool
	Schemas  []CustomFieldSchema
}

type CustomFieldValue struct {
	TicketID     uuid.UUID
	FieldID      uuid.UUID
	FieldName    string
	FieldLabel   string
	FieldType    string
	ValueText    *string
	ValueNumber  *float64
	ValueDate    *time.Time
	ValueBoolean *bool
	ValueUserID  *uuid.UUID
}

type CustomFieldValueInput struct {
	FieldID      uuid.UUID
	ValueText    *string
	ValueNumber  *float64
	ValueDate    *time.Time
	ValueBoolean *bool
	ValueUserID  *uuid.UUID
}

func (s *Store) ListCustomFields(ctx context.Context, projectID uuid.UUID) ([]CustomFieldDefinition, error) {
	fields, err := queryMany(ctx, s.db, mustSQL("custom_fields_by_project", nil), scanCustomField, projectID)
	if err != nil {
		return nil, err
	}
	schemas, err := s.listSchemasByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	// group schemas by fieldID
	schemaMap := make(map[uuid.UUID][]CustomFieldSchema)
	for _, sc := range schemas {
		schemaMap[sc.FieldID] = append(schemaMap[sc.FieldID], sc)
	}
	for i := range fields {
		fields[i].Schemas = schemaMap[fields[i].ID]
		if fields[i].Schemas == nil {
			fields[i].Schemas = []CustomFieldSchema{}
		}
	}
	return fields, nil
}

func (s *Store) GetCustomField(ctx context.Context, id, projectID uuid.UUID) (CustomFieldDefinition, error) {
	field, err := queryOne(ctx, s.db, mustSQL("custom_field_by_id", nil), scanCustomField, id, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CustomFieldDefinition{}, ErrCustomFieldNotFound
		}
		return CustomFieldDefinition{}, err
	}
	schemas, err := s.listSchemasByField(ctx, id)
	if err != nil {
		return CustomFieldDefinition{}, err
	}
	field.Schemas = schemas
	return field, nil
}

func (s *Store) CreateCustomField(ctx context.Context, projectID uuid.UUID, input CustomFieldCreateInput) (CustomFieldDefinition, error) {
	optJSON, err := json.Marshal(input.Options)
	if err != nil {
		return CustomFieldDefinition{}, err
	}
	field, err := queryOne(ctx, s.db, mustSQL("custom_field_insert", nil), scanCustomField,
		projectID, input.Name, input.Label, input.FieldType, string(optJSON), input.Required)
	if err != nil {
		return CustomFieldDefinition{}, err
	}
	if err := s.replaceSchemas(ctx, field.ID, input.Schemas); err != nil {
		return CustomFieldDefinition{}, err
	}
	field.Schemas = input.Schemas
	if field.Schemas == nil {
		field.Schemas = []CustomFieldSchema{}
	}
	return field, nil
}

func (s *Store) UpdateCustomField(ctx context.Context, id, projectID uuid.UUID, input CustomFieldUpdateInput) (CustomFieldDefinition, error) {
	optJSON, err := json.Marshal(input.Options)
	if err != nil {
		return CustomFieldDefinition{}, err
	}
	field, err := queryOne(ctx, s.db, mustSQL("custom_field_update", nil), scanCustomField,
		id, projectID, input.Label, string(optJSON), input.Required)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CustomFieldDefinition{}, ErrCustomFieldNotFound
		}
		return CustomFieldDefinition{}, err
	}
	if err := s.replaceSchemas(ctx, id, input.Schemas); err != nil {
		return CustomFieldDefinition{}, err
	}
	field.Schemas = input.Schemas
	if field.Schemas == nil {
		field.Schemas = []CustomFieldSchema{}
	}
	return field, nil
}

func (s *Store) DeleteCustomField(ctx context.Context, id, projectID uuid.UUID) error {
	return execOne(ctx, s.db, mustSQL("custom_field_delete", nil), ErrCustomFieldNotFound, id, projectID)
}

func (s *Store) GetCustomFieldValues(ctx context.Context, ticketID uuid.UUID) ([]CustomFieldValue, error) {
	return queryMany(ctx, s.db, mustSQL("ticket_custom_field_values_by_ticket", nil), scanCustomFieldValue, ticketID)
}

func (s *Store) SetCustomFieldValues(ctx context.Context, ticketID uuid.UUID, values []CustomFieldValueInput) ([]CustomFieldValue, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, mustSQL("ticket_custom_field_values_clear", nil), ticketID); err != nil {
		return nil, err
	}

	upsertSQL := mustSQL("ticket_custom_field_value_upsert", nil)
	for _, v := range values {
		var valueDate *time.Time
		if v.ValueDate != nil {
			t := *v.ValueDate
			valueDate = &t
		}
		if _, err := tx.Exec(ctx, upsertSQL,
			ticketID, v.FieldID,
			v.ValueText, v.ValueNumber, valueDate, v.ValueBoolean, v.ValueUserID,
		); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetCustomFieldValues(ctx, ticketID)
}

// CheckRequiredCustomFields verifies that for the given ticket type and target state,
// all required custom fields (per schema) have values in the provided values slice.
// Returns the names of any missing required fields.
func CheckRequiredCustomFields(
	fields []CustomFieldDefinition,
	values []CustomFieldValue,
	ticketType string,
	targetStateID uuid.UUID,
) []string {
	valMap := make(map[uuid.UUID]CustomFieldValue, len(values))
	for _, v := range values {
		valMap[v.FieldID] = v
	}

	var missing []string
	for _, f := range fields {
		for _, sc := range f.Schemas {
			if sc.TicketType != ticketType {
				continue
			}
			for _, reqState := range sc.RequiredStateIDs {
				if reqState == targetStateID {
					v, hasValue := valMap[f.ID]
					if !hasValue || isEmptyValue(v) {
						missing = append(missing, f.Label)
					}
					break
				}
			}
		}
	}
	return missing
}

func isEmptyValue(v CustomFieldValue) bool {
	return v.ValueText == nil &&
		v.ValueNumber == nil &&
		v.ValueDate == nil &&
		v.ValueBoolean == nil &&
		v.ValueUserID == nil
}

// --- internal helpers ---

func (s *Store) listSchemasByProject(ctx context.Context, projectID uuid.UUID) ([]CustomFieldSchema, error) {
	return queryMany(ctx, s.db, mustSQL("custom_field_schemas_by_project", nil), scanCustomFieldSchema, projectID)
}

func (s *Store) listSchemasByField(ctx context.Context, fieldID uuid.UUID) ([]CustomFieldSchema, error) {
	return queryMany(ctx, s.db, mustSQL("custom_field_schemas_by_field", nil), scanCustomFieldSchema, fieldID)
}

func (s *Store) replaceSchemas(ctx context.Context, fieldID uuid.UUID, schemas []CustomFieldSchema) error {
	if _, err := s.db.Exec(ctx, mustSQL("custom_field_schemas_delete_by_field", nil), fieldID); err != nil {
		return err
	}
	upsertSQL := mustSQL("custom_field_schema_upsert", nil)
	for _, sc := range schemas {
		ids, err := json.Marshal(sc.RequiredStateIDs)
		if err != nil {
			return err
		}
		if _, err := s.db.Exec(ctx, upsertSQL, fieldID, sc.TicketType, string(ids)); err != nil {
			return err
		}
	}
	return nil
}

func scanCustomField(row pgx.Row) (CustomFieldDefinition, error) {
	var f CustomFieldDefinition
	var optionsJSON []byte
	err := row.Scan(
		&f.ID, &f.ProjectID, &f.Name, &f.Label, &f.FieldType,
		&optionsJSON, &f.Order, &f.Required, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return CustomFieldDefinition{}, err
	}
	if len(optionsJSON) > 0 {
		if err := json.Unmarshal(optionsJSON, &f.Options); err != nil {
			return CustomFieldDefinition{}, err
		}
	}
	if f.Options == nil {
		f.Options = []CustomFieldOption{}
	}
	return f, nil
}

func scanCustomFieldSchema(row pgx.Row) (CustomFieldSchema, error) {
	var sc CustomFieldSchema
	var idsJSON []byte
	err := row.Scan(&sc.FieldID, &sc.TicketType, &idsJSON)
	if err != nil {
		return CustomFieldSchema{}, err
	}
	if len(idsJSON) > 0 {
		if err := json.Unmarshal(idsJSON, &sc.RequiredStateIDs); err != nil {
			return CustomFieldSchema{}, err
		}
	}
	if sc.RequiredStateIDs == nil {
		sc.RequiredStateIDs = []uuid.UUID{}
	}
	return sc, nil
}

func scanCustomFieldValue(row pgx.Row) (CustomFieldValue, error) {
	var v CustomFieldValue
	err := row.Scan(
		&v.TicketID, &v.FieldID, &v.FieldName, &v.FieldLabel, &v.FieldType,
		&v.ValueText, &v.ValueNumber, &v.ValueDate, &v.ValueBoolean, &v.ValueUserID,
	)
	return v, err
}
