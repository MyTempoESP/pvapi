---
to: models/<%= h.changeCase.no(name) %>model/blueprint.go
---
package <%= h.changeCase.no(name) %>model

type <%= h.changeCase.pascal(name) %>Blueprint struct {
    Field any `json:"field"`
}
