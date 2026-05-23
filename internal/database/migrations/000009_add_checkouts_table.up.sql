CREATE TABLE IF NOT EXISTS public.checkouts
(
	id                 serial         not null,
	cart_id            integer        not null,
	consignment_id     integer                , -- nullable
	unique_id          uuid           default uuidv4(),
	value              numeric(12, 2) not null,
	status             text           not null,
	description        text           not null,
	external_reference text           not null,
	metadata           jsonb                  ,
	processed_at       timestamp      with time zone default now(),
	primary key(id)
);

ALTER TABLE IF EXISTS public.checkouts
    ADD CONSTRAINT checkout_cart_relation FOREIGN KEY (cart_id)
    REFERENCES public.carts (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;
CREATE INDEX IF NOT EXISTS fki_payments_cart
    ON public.checkouts(cart_id);
