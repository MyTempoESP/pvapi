CREATE OR REPLACE VIEW consignments_active AS
SELECT *
FROM consignments
WHERE status NOT IN ('ARQUIVADO', 'RASCUNHO');
