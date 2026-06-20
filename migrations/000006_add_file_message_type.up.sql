ALTER TABLE messages
DROP CONSTRAINT IF EXISTS messages_type_valid;

ALTER TABLE messages
ADD CONSTRAINT messages_type_valid CHECK (type IN ('text', 'image', 'system', 'file'));