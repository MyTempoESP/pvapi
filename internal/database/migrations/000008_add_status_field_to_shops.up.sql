ALTER TABLE public.shops ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'active';
UPDATE public.shops SET status = 'active' WHERE status IS NULL;

CREATE OR REPLACE VIEW public.shops_info_full_view AS
SELECT
    s.id AS shop_id,
    s.name AS name,
    s.slug AS slug,
    s.telephone AS telephone,
    s.status AS status,

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

