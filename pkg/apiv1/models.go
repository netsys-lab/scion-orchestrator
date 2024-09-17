package apiv1

// Define a struct to represent the incoming JSON data
type ApiCSR struct {
	Subject Subject `json:"subject" binding:"required"` // "binding:required" ensures this field must be present
}

// Subject represents the nested structure of the incoming JSON
type Subject struct {
	ISDAS              string `json:"isd_as" binding:"required"`
	CommonName         string `json:"common_name" binding:"required"`
	Country            string `json:"country" `
	Locality           string `json:"locality"`
	Organization       string `json:"organization" `
	OrganizationalUnit string `json:"organizational_unit" `
	PostalCode         string `json:"postal_code" `
	Province           string `json:"province" `
	SerialNumber       string `json:"serial_number" `
	StreetAddress      string `json:"street_address"`
}
