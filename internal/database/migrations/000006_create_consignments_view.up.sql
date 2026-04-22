CREATE OR REPLACE VIEW consignments_listing_view AS
SELECT
    c.id              AS consignment_id,
    c.comission_type  AS consignment_comission_type,
    c.comission_value AS consignment_comission_value,
    c.status          AS consignment_status,
    c.created_at      AS consignment_created_at,
    c.slug            AS consignment_slug,
    c.code            AS consignment_code,

    m.consignee_id AS consignee_id,
    m.user_name    AS consignee_name,

    s.shop_id AS shop_id,
    s.name    AS shop_name,

    p.id   AS pickup_location_id,
    p.name AS pickup_location_name
FROM consignments c
LEFT JOIN shops_info_view      s USING (shop_id)
LEFT JOIN consignees_info_view m USING (consignee_id)
LEFT JOIN pickup_locations     p ON p.id = c.pickup_location_id
;
