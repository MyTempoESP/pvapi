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

CREATE OR REPLACE VIEW aliased_addresses_view AS
SELECT
  a.id AS address_id,
  a.street AS addrstreet,
  a.number AS addrnumber,
  a.neighborhood AS addrneigh,
  a.city AS addrcity,
  a.state AS addrstate,
  a.cep AS addrcep,
  a.complement AS addrcomplement
FROM addresses a;
