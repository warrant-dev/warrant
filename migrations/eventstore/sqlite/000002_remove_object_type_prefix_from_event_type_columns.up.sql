UPDATE accessEvent
SET type = SUBSTR(type, INSTR(type, '.') + 1);

UPDATE resourceEvent
SET type = SUBSTR(type, INSTR(type, '.') + 1);
