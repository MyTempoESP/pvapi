---
to: models/<%= h.changeCase.no(name) %>model/resources.go
---
<%
    Flat = h.changeCase.no(name)
    Class = h.changeCase.pascal(name)
    Model = Flat + "model"
    Char = name.charAt(0)
%>package <%= Model %>

type <%= Class %>Resource struct {
    Field any `json:"field"`
}

func (<%= Char %> <%= Class %>) ToResource() <%= Class %>Resource {
    return <%= Class %>Resource{
        Field: <%= Char %>.Field,
    }
}
