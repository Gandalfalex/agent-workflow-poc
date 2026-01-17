CREATE SEQUENCE IF NOT EXISTS ticket_number_seq START WITH 1000;

ALTER TABLE tickets
  ADD COLUMN IF NOT EXISTS number bigint NOT NULL DEFAULT nextval('ticket_number_seq');

ALTER TABLE tickets
  ADD CONSTRAINT tickets_number_unique UNIQUE (number);
