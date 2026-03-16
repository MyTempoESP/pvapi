## Basic controller template

```go
package template

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ListHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {

	}
}

func CreateHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {

	}
}

func ViewHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {

	}
}

func UpdateHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {

	}
}
```

## Model template

```go
package templatemodel 

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Template struct {
	ID int `json:"id" db:"template_id"`
}

func Update(
	ctx context.Context,
	pool *pgxpool.Pool,
	templateID int,
	template UpdateTemplateRequest,
) (err error) {
	return
}

func Create(
	ctx context.Context,
	pool *pgxpool.Pool,
	template CreateTemplateRequest,
) (templateID int, err error) {
	return
}

func Get(
	ctx context.Context,
	pool *pgxpool.Pool,
) (templateCollection []TemplateStruct, err error) {
	return
}

func Find(
	ctx context.Context,
	pool *pgxpool.Pool,
	id string,
) (template TemplateStruct, err error) {
	return
}
```

## Request template

```go
package templatemodel

type CreateTemplateRequest struct {
}

type UpdateTemplateRequest struct {
}
```

## Resource template

```go
package templatemodel

type TemplateResource struct {
	ID int `json:"id"`
}

func (template Template) ToResource() TemplateResource {
	return TemplateResource{
		ID: template.ID,
	}
}
```
