package gate

import (
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/gate/roles"
	"linhazero.com/pvapi/models/consigneemodel"
	"linhazero.com/pvapi/models/policymodel"
	"linhazero.com/pvapi/models/shopmodel"
)

type ReportGate struct {
	Base db.Queryable
}

func (p *ReportGate) InvolvedIn(u User, r Resource) (bool, error) {
	id := r.Metadata["id"]

	if u.Role == roles.Shop {
		return shopmodel.HasReport(
			r.Context, p.Base,
			u.ActiveRoleID, id,
		)
	}

	if u.Role == roles.Consignee {
		return consigneemodel.HasReport(
			r.Context, p.Base,
			u.ActiveRoleID, id,
		)
	}

	return false, nil
}

func (p *ReportGate) Owns(u User, r Resource) (bool, error) {
	if u.Role != roles.Shop {
		return false, nil
	}

	id := r.Metadata["id"]

	return shopmodel.HasReport(
		r.Context, p.Base,
		u.ActiveRoleID, id,
	)
}

func (report *ReportGate) Dispatch(
	u User,
	r Resource,
) (bool, error) {
	id := r.Params.ByName("report")
	r.Metadata["id"] = id

	if r.Method == "GET" {
		if id == "" {
			return report.List(u, r)
		}
		return report.View(u, r)
	}

	if r.Method == "POST" {
		return report.Create(u, r)
	}

	if r.Method == "PUT" {
		return report.Modify(u, r)
	}

	if r.Method == "DELETE" {
		return report.Delete(u, r)
	}

	return false, nil
}

func (report *ReportGate) Create(u User, r Resource) (bool, error) {
	return u.Role == roles.Shop || u.Role == roles.Consignee, nil
}

func (report *ReportGate) View(u User, r Resource) (bool, error) {
	return report.InvolvedIn(u, r)
}

func (report *ReportGate) List(u User, r Resource) (bool, error) {
	return u.Role == roles.Shop || u.Role == roles.Consignee, nil
}

func (report *ReportGate) Modify(u User, r Resource) (bool, error) {
	return report.InvolvedIn(u, r)
}

func (report *ReportGate) Delete(u User, r Resource) (bool, error) {
	return report.Owns(u, r)
}

func NewReport(base db.Queryable) policymodel.Gate {
	return &ReportGate{
		Base: base,
	}
}
