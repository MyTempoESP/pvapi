CREATE OR REPLACE FUNCTION insert_new_address(
    a_street       TEXT,
    a_number       INTEGER,
    a_neighborhood TEXT,
    a_city         TEXT,
    a_state        TEXT,
    a_cep          TEXT,
    a_complement   TEXT
)
RETURNS INTEGER AS $$
    INSERT INTO addresses
		(
			street,
			number,
			neighborhood,
			city,
			state,
			cep,
			complement
		)
	VALUES
		(
            a_street,
            a_number,
            a_neighborhood,
            a_city,
            a_state,
            a_cep,
            a_complement
		)
	RETURNING id
$$ LANGUAGE sql VOLATILE;
