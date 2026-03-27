---
to: gate/<%= h.changeCase.no(name) %>.go
---
<%
    Flat = h.changeCase.no(name)
    Table = h.inflection.pluralize(Flat)

    Class = h.changeCase.pascal(name)
    Model = Flat + "model"

    Char = name.charAt(0)
%>package gate

import (
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/policymodel"
)

type <%= Class %>Gate struct {
	Base db.Queryable
}

func (<%= Char %> *<%= Class %>Gate) Dispatch(
	u User,
	r Resource,
) (bool, error) {
	id := r.Params.ByName("<%= Flat %>")
	r.Metadata["id"] = id

	if r.Method == "GET" {
		if id == "" {
			return <%= Char %>.List(u, r)
		}
		return <%= Char %>.View(u, r)
	}

	if r.Method == "PUT" {
		return <%= Char %>.Modify(u, r)
	}

	if r.Method == "DELETE" {
		return <%= Char %>.Delete(u, r)
	}

	return false, nil
}

func (<%= Char %> *<%= Class %>Gate) Create(u User, r Resource) (bool, error) {
	return false, nil
}

func (<%= Char %> *<%= Class %>Gate) View(u User, r Resource) (bool, error) {
	return false, nil
}

func (<%= Char %> *<%= Class %>Gate) List(u User, r Resource) (bool, error) {
	return false, nil
}

func (<%= Char %> *<%= Class %>Gate) Modify(u User, r Resource) (bool, error) {
	return false, nil
}

func (<%= Char %> *<%= Class %>Gate) Delete(u User, r Resource) (bool, error) {
	return false, nil
}

func New<%= Class %>(base db.Queryable) policymodel.Gate {
	return &<%= Class %>Gate{
		Base: base,
	}
}
