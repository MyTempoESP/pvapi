CREATE OR REPLACE VIEW consignments_listing_view AS
SELECT
    c.id                 AS consignment_id,
    c.comission_type     AS consignment_comission_type,
    c.comission_value    AS consignment_comission_value,
    c.pickup_location_id AS consignment_pickup_location,
    c.status             AS consignment_status,

    m.consignee_id AS consignee_id,
    m.user_name    AS consignee_name,

    s.shop_id   AS shop_id,
    s.name      AS shop_name
FROM consignments c
LEFT JOIN shops_info_view      s USING (shop_id)
LEFT JOIN consignees_info_view m USING (consignee_id)
;
