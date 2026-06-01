BEGIN;

ALTER TABLE IF EXISTS public.checkout_shop
    ADD COLUMN IF NOT EXISTS client_id integer;

ALTER TABLE IF EXISTS public.checkout_shop
    ADD CONSTRAINT checkout_shop_client_relation FOREIGN KEY (client_id)
    REFERENCES public.clients (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;

COMMIT;
