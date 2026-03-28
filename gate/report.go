package gate

import (
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/policymodel"
)

type ReportGate struct {
	Base db.Queryable
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

	if r.Method == "PUT" {
		return report.Modify(u, r)
	}

	if r.Method == "DELETE" {
		return report.Delete(u, r)
	}

	return false, nil
}

func (report *ReportGate) Create(u User, r Resource) (bool, error) {
	return false, nil
}

func (report *ReportGate) View(u User, r Resource) (bool, error) {
	return false, nil
}

func (report *ReportGate) List(u User, r Resource) (bool, error) {
	return false, nil
}

func (report *ReportGate) Modify(u User, r Resource) (bool, error) {
	return false, nil
}

func (report *ReportGate) Delete(u User, r Resource) (bool, error) {
	return false, nil
}

func NewReport(base db.Queryable) policymodel.Gate {
	return &ReportGate{
		Base: base,
	}
}
