-- Add WIP limit fields to workflow_states
ALTER TABLE workflow_states ADD COLUMN wip_limit int;
ALTER TABLE workflow_states ADD COLUMN wip_enforcement boolean NOT NULL DEFAULT false;
