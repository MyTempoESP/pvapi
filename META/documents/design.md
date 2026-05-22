# Design notes

## Product Model

the Product model has a novel approach to simplify image/description handling,
to describe a product, i.e., create a description, you will pass over a
"description" field on the request. Such request is handled in the following steps:

1. Grab the string
2. Create a default "description"->text specification
3. Associate any images using specification_images table

### ultimate BS take on data modeling (did not age well):

```
### "Everything is a specification" pattern

Every single piece of info on a product is a specification, this includes,
but is not limited to:

- The product description
- The product banner image
- The product delivery images
- The product category
- The product colors

This gets fetched from the DB as some special addons for each product.

e.g. when i need some product to have color selection i must create a new specification
with the available colors as idk a separated field like #FF2211, #00F222, #0022F2
```


### note from the future:

The last section was dumb, i'm leaving it here as a wall of shame.
