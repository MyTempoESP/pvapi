CREATE OR REPLACE VIEW public.shops_info_full_view AS
SELECT
    s.id AS shop_id,
    s.name AS name,
    s.slug AS slug,
    s.telephone AS telephone,
    'active' AS status,

    u.id AS user_id,
    u.name AS user_name,
    u.email AS user_email,

    a.street AS addrstreet,
    a.number AS addrnumber,
    a.neighborhood AS addrneigh,
    a.city AS addrcity,
    a.state AS addrstate,
    a.cep AS addrcep,
    a.complement AS addrcomplement
FROM public.shops s
LEFT JOIN public.users u ON u.id = s.user_id
LEFT JOIN public.addresses a ON a.id = s.address_id;

CREATE OR REPLACE VIEW public.shops_info_view AS
SELECT
    shop_id,
    name,
    slug,
    telephone,
    status,

    -- ommitted user_id
    user_name,
    user_email,

    addrstreet,
    addrnumber,
    addrneigh,
    addrcity,
    addrstate,
    addrcep,
    addrcomplement
FROM public.shops_info_full_view;

CREATE OR REPLACE VIEW public.shops_info_detailed_view AS
SELECT
    v.*,

    COALESCE((
        SELECT COUNT(*)
        FROM products p
        WHERE p.shop_id = v.shop_id
    ), 0) AS products_count,

    COALESCE((
        SELECT COUNT(*)
        FROM pickup_locations pl
        WHERE pl.shop_id = v.shop_id
    ), 0) AS pickups
FROM public.shops_info_full_view v;
