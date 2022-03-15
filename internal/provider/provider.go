package mira

import (
	"context"
	// "os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"terraform-provider-mira/miraclient"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				// "url": &schema.Schema{
				// 	Type:        schema.TypeString,
				// 	Optional:    false,
				// 	// DefaultFunc: schema.EnvDefaultFunc("MIRA_ADDRESS", nil),
				// 	DefaultFunc: schema.EnvDefaultFunc(os.Getenv("MIRA_ADDRESS"), nil),
				// },
				// "username": &schema.Schema{
				// 	Type:        schema.TypeString,
				// 	Optional:    true,
				// 	DefaultFunc: schema.EnvDefaultFunc(os.Getenv("MIRA_USERNAME"), nil),
				// },
				// "password": &schema.Schema{
				// 	Type:        schema.TypeString,
				// 	Optional:    true,
				// 	DefaultFunc: schema.EnvDefaultFunc(os.Getenv("MIRA_PASSWORD"), nil),
				// },
			},
			DataSourcesMap: map[string]*schema.Resource{
				"mira_available_subnet_data_source": dataSourceMiraAvailableSubnets(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"mira_allocated_subnet_resource": resourceMiraAllocatedSubnet(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

// NOT NEEDED - I am instantiating the client struct from the miraclient package,
//              and I am pulling user creds from the environment in the new client
//              so this data structure is specified in miraclient and not needed here

// type apiClient struct {
	// Add whatever fields, client or connection info, etc. here
	// you would need to setup to communicate with the upstream
	// API.
//	Username string
//	Password string
//	Uri string
//	UserAgent string
//}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
		// create new client from miraclient package, 
		// the new client does not take any variables,
		// it uses os.Getenv to collect the credentials 
		// and url from the environment that terraform instantiates
		apiClient, err := miraclient.NewClient()
		if err != nil {
			return nil, diag.FromErr(err)
		}
		return apiClient, nil
	}
}
