BEGIN;

CREATE OR REPLACE VIEW public.consignments_listing_view AS
 SELECT c.id AS consignment_id,
    c.comission_type AS consignment_comission_type,
    c.comission_value AS consignment_comission_value,
    c.status AS consignment_status,
    c.created_at AS consignment_created_at,
    c.slug AS consignment_slug,
    c.code AS consignment_code,
    m.consignee_id,
    m.user_name AS consignee_name,
    s.shop_id,
    s.name AS shop_name
   FROM ((public.consignments c
     LEFT JOIN public.shops_info_view s USING (shop_id))
     LEFT JOIN public.consignees_info_view m USING (consignee_id));

ALTER TABLE IF EXISTS public.consignments
    DROP COLUMN IF EXISTS name;

COMMIT;
