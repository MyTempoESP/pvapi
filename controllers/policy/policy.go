package policy

import (
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/gate"
	"linhazero.com/pvapi/models/policymodel"
)

func New(base db.Queryable) *policymodel.Bouncer {

	bouncer := policymodel.NewBouncer(base)

	bouncer.Register(policymodel.Guard{
		Path: "/estabelecimentos",
		Gate: gate.NewShop(base),
	})

	bouncer.Register(policymodel.Guard{
		Path: "/monitores",
		Gate: gate.NewConsignee(base),
	})

	bouncer.Register(policymodel.Guard{
		Path: "/produtos",
		Gate: gate.NewProduct(base),
	})

	bouncer.Register(policymodel.Guard{
		Path: "/pontos",
		Gate: gate.NewPickups(base),
	})

	bouncer.Register(policymodel.Guard{
		Path: "/remessas",
		Gate: gate.NewConsignment(base),
	})

	bouncer.Register(policymodel.Guard{
		Path: "/ocorrencias",
		Gate: gate.NewReport(base),
	})

	bouncer.Register(policymodel.Guard{
		Path: "/admin",
		Gate: gate.NewAdmin(base),
	})

	bouncer.Register(policymodel.Guard{
		Path: "/clientes",
		Gate: gate.NewClient(base),
	})

	bouncer.Register(policymodel.Guard{
		Path: "/clients",
		Gate: gate.NewClient(base),
	})

	bouncer.Register(policymodel.Guard{
		Path: "/categorias",
		Gate: gate.NewCategory(base),
	})

	bouncer.Register(policymodel.Guard{
		Path: "/usuario",
		Gate: gate.NewUsuario(),
	})

	return bouncer
}
