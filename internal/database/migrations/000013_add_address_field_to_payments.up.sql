ALTER TABLE IF EXISTS public.payments
    ADD COLUMN IF NOT EXISTS address_id integer;

ALTER TABLE IF EXISTS public.payments
    ADD CONSTRAINT fk_payments_address
    FOREIGN KEY (address_id)
    REFERENCES public.addresses (id)
    MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE SET NULL
    NOT VALID;
