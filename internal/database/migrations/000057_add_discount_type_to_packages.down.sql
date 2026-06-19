DROP FUNCTION IF EXISTS effective_unit_price(numeric, numeric, text);

ALTER TABLE packages
	DROP CONSTRAINT IF EXISTS discount_range;

ALTER TABLE packages
	ADD CONSTRAINT discount_range CHECK (
		discount >= 0 AND discount <= 100
	);

ALTER TABLE packages
	DROP COLUMN IF EXISTS discount_type;
