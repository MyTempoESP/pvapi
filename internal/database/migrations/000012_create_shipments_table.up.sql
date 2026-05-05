CREATE TABLE IF NOT EXISTS public.shipments
(
    id SERIAL PRIMARY KEY,
    address_id INTEGER,
    payment_id INTEGER NOT NULL,
    client_id INTEGER NOT NULL,
    total_amount NUMERIC(12, 2) NOT NULL,
    code UUID NOT NULL DEFAULT uuidv4(),
    status TEXT NOT NULL,
    shop_id INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_shipment_code UNIQUE (code)
);

ALTER TABLE IF EXISTS public.shipments
    ADD CONSTRAINT shipment_address_relation FOREIGN KEY (address_id)
    REFERENCES public.addresses (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE SET NULL;

ALTER TABLE IF EXISTS public.shipments
    ADD CONSTRAINT shipment_payment_relation FOREIGN KEY (payment_id)
    REFERENCES public.payments (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;

ALTER TABLE IF EXISTS public.shipments
    ADD CONSTRAINT shipment_client_relation FOREIGN KEY (client_id)
    REFERENCES public.clients (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;

ALTER TABLE IF EXISTS public.shipments
    ADD CONSTRAINT shipment_shop_relation FOREIGN KEY (shop_id)
    REFERENCES public.shops (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;

