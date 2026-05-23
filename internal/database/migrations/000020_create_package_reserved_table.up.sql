BEGIN;

CREATE TABLE IF NOT EXISTS public.package_reserved
(
    id               serial         NOT NULL,
    package_id       integer        NOT NULL,
    checkout_id      integer        NOT NULL,
    quantity         integer        NOT NULL CHECK (quantity > 0),
    reserved_at      timestamp      with time zone DEFAULT now(),
    PRIMARY KEY (id)
);

ALTER TABLE IF EXISTS public.package_reserved
    ADD CONSTRAINT package_reserved_package_relation FOREIGN KEY (package_id)
    REFERENCES public.packages (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;
CREATE INDEX IF NOT EXISTS fki_package_reserved_package
    ON public.package_reserved(package_id);

ALTER TABLE IF EXISTS public.package_reserved
    ADD CONSTRAINT package_reserved_checkout_relation FOREIGN KEY (checkout_id)
    REFERENCES public.checkouts (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;
CREATE INDEX IF NOT EXISTS fki_package_reserved_checkout
    ON public.package_reserved(checkout_id);

COMMIT;
