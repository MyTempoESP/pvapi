package productmodel

import (
	"mime/multipart"

	"github.com/shopspring/decimal"
)

/* before preparation:
 * {
 * 	'name'       , string
 * 	'price'      , decimal
 * 	'stock'      , integer
 * 	'discount'   , decimal
 * 	'description', string
 * 	'category'   , string
 * 	'featured'   , boolean
 *
 * 	'imageUrl'           , multipart form-data
 * 	'descriptionImages'  , multipart form-data
 * 	'specificationImages', multipart form-data
 * 	'deliveryImages'     , multipart form-data
 *
 * 	'specifications'           , []{'name':string, 'value':string}
 * 	'allowCustomName'          , boolean
 * 	'allowCustomModality'      , boolean
 * 	'allowCustomColorSelection', boolean
 * 	'colors'                   , []string
 * }
 */

/* after preparation (blueprint)
 * {
 * 	'uuid'       , UUID   (generated)
 * 	'code'       , string (generated)
 *  	'name'       , string
 * 	'price'      , decimal
 * 	'stock'      , integer
 * 	'discount'   , decimal
 * 	'category'   , string
 * 	'featured'   , boolean
 *
 * 	'image'      ,    string (full path)  (generated)
 * 	'desc_images', [](string (full path)) (generated
 * 	'pack_images', [](string (full path)) (generated
 * 	'spec_images', [](string (full path)) (generated
 *
 * 	'specifications', []{'name':string, 'value':string}
 * 	 [0]:	'category'   , {'name':string, 'value':string} ('value' = generated slug),
 * 	 [1]:	'description', {'name':string, 'value':string}
 * 	 [n]:   ...
 *
 * 	'policy.colors'        , {'name':string,'value':[]string,'enabled':boolean}
 * 	'policy.customName'    , {'name':string,'value':null    ,'enabled':boolean}
 * 	'policy.customModality', {'name':string,'value':null    ,'enabled':boolean}
 * 	'policy.customColor'   , {'name':string,'value':null    ,'enabled':boolean}
 * }
 */

/* request example
------geckoformboundary8361523800a4ba83aea6b83cd92ae156
Content-Disposition: form-data; name="name"

aaaaaa
------geckoformboundary8361523800a4ba83aea6b83cd92ae156
Content-Disposition: form-data; name="description"

aaa
------geckoformboundary8361523800a4ba83aea6b83cd92ae156
Content-Disposition: form-data; name="price"

2
------geckoformboundary8361523800a4ba83aea6b83cd92ae156
Content-Disposition: form-data; name="category"

porta-medalhas
------geckoformboundary8361523800a4ba83aea6b83cd92ae156
Content-Disposition: form-data; name="stock"

1
------geckoformboundary8361523800a4ba83aea6b83cd92ae156
Content-Disposition: form-data; name="featured"

true
------geckoformboundary8361523800a4ba83aea6b83cd92ae156
Content-Disposition: form-data; name="discount"

12
------geckoformboundary8361523800a4ba83aea6b83cd92ae156
Content-Disposition: form-data; name="allowCustomization"

false
------geckoformboundary8361523800a4ba83aea6b83cd92ae156
Content-Disposition: form-data; name="allowCustomName"

false
------geckoformboundary8361523800a4ba83aea6b83cd92ae156
Content-Disposition: form-data; name="allowCustomModality"

false
------geckoformboundary8361523800a4ba83aea6b83cd92ae156
Content-Disposition: form-data; name="allowCustomColorSelection"

false
------geckoformboundary8361523800a4ba83aea6b83cd92ae156
Content-Disposition: form-data; name="colors"

["#000000"]
------geckoformboundary8361523800a4ba83aea6b83cd92ae156
Content-Disposition: form-data; name="specifications"

[{"name":"a","value":"b"}]
------geckoformboundary8361523800a4ba83aea6b83cd92ae156
Content-Disposition: form-data; name="imageUrl"

https://.com
------geckoformboundary8361523800a4ba83aea6b83cd92ae156--
*/

type CreateProductRequest struct {
	Name        string          `form:"name"        binding:"required"`
	Price       string          `form:"price"       binding:"required"`
	Stock       int             `form:"stock"       binding:"required"`
	Discount    decimal.Decimal `form:"discount"    binding:"required"`
	Description string          `form:"description" binding:"required"`
	Category    string          `form:"category"    binding:"required"`
	Featured    *bool           `form:"featured"`

	Image      *multipart.FileHeader   `form:"imageUrl" binding:"required"`
	DescImages []*multipart.FileHeader `form:"descriptionImages"`
	SpecImages []*multipart.FileHeader `form:"specificationImages"`
	PackImages []*multipart.FileHeader `form:"deliveryImages"`

	SpecificationsJSON string `form:"specifications" binding:"required"`

	// Policies
	ColorListJSON        string `form:"colors"`
	CustomName           bool   `form:"allowCustomName"`
	customModality       bool   `form:"allowCustomModality"`
	CustomColorSelection bool   `form:"allowCustomColorSelection"`
}
