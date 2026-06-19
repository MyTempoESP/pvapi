-- ponytail: effective_unit_price is a single function instead of duplicating the
-- percentage-vs-fixed branch in every checkout expression. Add when another
-- discount type is needed — at that point the function signature changes anyway.

ALTER TABLE packages
	ADD COLUMN discount_type text NOT NULL DEFAULT 'percentage';

ALTER TABLE packages
	DROP CONSTRAINT IF EXISTS discount_range;

ALTER TABLE packages
	ADD CONSTRAINT discount_range CHECK (
		discount >= 0 AND (
			discount_type = 'fixed' OR discount <= 100
		)
	);

CREATE OR REPLACE FUNCTION effective_unit_price(
	price numeric,
	discount numeric,
	discount_type text
)
RETURNS numeric
LANGUAGE SQL
IMMUTABLE
AS $$
	SELECT CASE
		WHEN discount_type = 'fixed' THEN GREATEST(price - discount, 0)
		ELSE price * (1.0 - discount / 100.0)
	END
$$;
