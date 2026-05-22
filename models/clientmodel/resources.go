package clientmodel

/*
   'id' => (string) $client->id,
   'name' => $client->name,
   'email' => $client->email,
   'phone' => $client->phone,
*/

type ClientResource struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

func (c Client) ToResource() ClientResource {
	return ClientResource(c)
}
