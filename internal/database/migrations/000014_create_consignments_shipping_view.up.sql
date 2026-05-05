CREATE OR REPLACE VIEW consignment_shipment_view AS
SELECT 
    c.id AS consignment_id,
    c.comission_type AS consignment_comission_type,
    c.comission_value::numeric AS consignment_comission_value,
    c.status AS consignment_status,
    c.created_at AS consignment_created_at,
    c.slug AS consignment_slug,
    c.code::text AS consignment_code,

    co.id AS consignee_id,
    u.name AS consignee_name,

    s.id AS shop_id,
    s.name AS shop_name,

    pl.id AS pickup_location_id,
    pl.name AS pickup_location_name,

    'consignment' AS record_type

FROM consignments c
LEFT JOIN consignees co ON co.id = c.consignee_id
LEFT JOIN users u ON u.id = co.user_id
JOIN shops s ON s.id = c.shop_id
LEFT JOIN pickup_locations pl ON pl.id = c.pickup_location_id

UNION ALL

SELECT
    sh.id AS consignment_id,
    NULL AS consignment_comission_type,
    NULL AS consignment_comission_value,
    sh.status AS consignment_status,
    sh.created_at AS consignment_created_at,
    'shipment' AS consignment_slug,
    sh.code::text AS consignment_code,

    NULL AS consignee_id,
    NULL AS consignee_name,

    sh.shop_id,
    s.name AS shop_name,

    NULL AS pickup_location_id,
    COALESCE(a.street || ', ' || a.number, a.street, a.number, 'Endereço de entrega') AS pickup_location_name,

    'shipment' AS record_type

FROM shipments sh
JOIN shops s ON s.id = sh.shop_id
LEFT JOIN addresses a ON a.id = sh.address_id;
