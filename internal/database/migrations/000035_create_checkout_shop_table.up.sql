BEGIN;

CREATE TABLE IF NOT EXISTS public.checkout_shop
(
    checkout_id        integer        NOT NULL,
    shop_id            integer        NOT NULL,
    checkout_unique_id uuid           NOT NULL,
    shop_value         numeric(12, 2) NOT NULL,
    shop_tax           numeric(12, 2) NOT NULL,
    shop_commission    numeric(12, 2) NOT NULL,
    created_at         timestamp with time zone DEFAULT now(),
    PRIMARY KEY (checkout_id, shop_id)
);

ALTER TABLE IF EXISTS public.checkout_shop
    ADD CONSTRAINT checkout_shop_checkout_relation FOREIGN KEY (checkout_id)
    REFERENCES public.checkouts (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;

ALTER TABLE IF EXISTS public.checkout_shop
    ADD CONSTRAINT checkout_shop_shop_relation FOREIGN KEY (shop_id)
    REFERENCES public.shops (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;

ALTER TABLE IF EXISTS public.checkouts
    ADD COLUMN IF NOT EXISTS value_discrepancy numeric(12, 2) NOT NULL DEFAULT 0;

COMMIT;
