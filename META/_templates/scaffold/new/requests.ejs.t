---
to: models/<%= h.changeCase.no(name) %>model/requests.go
---
<%
    Flat = h.changeCase.no(name)
    table = h.inflection.pluralize(Flat)

    Class = h.changeCase.pascal(name)
    Model = Flat + "model"

    Char = name.charAt(0)
%>package <%= Model %>

type Create<%= Class %>Request struct {
    Field any `json:"field" binding:"required"`
}

type Update<%= Class %>Request struct {
    Field any `json:"field"`
}

func (<%= Char %> Create<%= Class %>Request) Prepare() <%= Class %>Blueprint {
    return <%= Class %>Blueprint{
        Field: <%= Char %>.Field,
    }
}

func (<%= Char %> Update<%= Class %>Request) Prepare(id int) <%= Class %> {
    return <%= Class %>{
        ID:    id,
        Field: <%= Char %>.Field,
    }
}
