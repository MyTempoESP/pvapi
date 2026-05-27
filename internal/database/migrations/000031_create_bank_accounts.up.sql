BEGIN;

CREATE TABLE IF NOT EXISTS public.bank_accounts
(
    id               serial       NOT NULL,
    shop_id          integer      NOT NULL,
    name             text         NOT NULL,
    bank_name        text         NOT NULL DEFAULT '',
    pix_key          text,
    pix_titular      text,
    transfer_details jsonb,
    created_at       timestamp with time zone DEFAULT now(),
    PRIMARY KEY (id)
);

ALTER TABLE IF EXISTS public.bank_accounts
    ADD CONSTRAINT bank_accounts_shop_relation FOREIGN KEY (shop_id)
    REFERENCES public.shops (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE CASCADE
    NOT VALID;

COMMIT;
