CREATE TABLE IF NOT EXISTS public.payments
(
	id                 serial         not null,
	cart_id            integer        not null,
	transaction_amount numeric(12, 2) not null,
	quantity           integer        not null,
	amount             numeric(12, 2) not null,
	status             text           not null,
	description        text           not null,
	title              text           not null,
	payment_id         text           not null,
	payer_name         text           not null,
	payer_email        text           not null,
	payer_surname      text           not null,
	payment_method     text           not null,
	currency           text           not null,
	external_reference text           not null,
	method             text           not null,
	metadata           jsonb                  ,
	processed_at       timestamp      with time zone default now(),
	primary key(id)
);

ALTER TABLE IF EXISTS public.payments
    ADD CONSTRAINT payment_cart_relation FOREIGN KEY (cart_id)
    REFERENCES public.carts (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;
CREATE INDEX IF NOT EXISTS fki_a
    ON public.payments(cart_id);
