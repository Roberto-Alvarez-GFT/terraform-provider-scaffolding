package mira

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	// the db mira client
	"terraform-provider-mira/miraclient"
)

// ======================================================================
// THE FUNCTION ASSIGNED TO THE DATASOURCE IN provider.gp [RETURN SCHEMA]
// ======================================================================

func dataSourceMiraAvailableSubnets() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "A data source in the Terraform provider Mira for listing available subnets from with a specified address range.",

		// the function in this file assigned to the READ CRUD call
		ReadContext: dataSourceMiraAvailableSubnetsRead,

		// set the resource fields in the schema
		Schema: map[string]*schema.Schema{
			// these two resources are populated via git and terraform
			"requestrange": {
				// This description is used by the documentation generator and the language server.
				Description: "The Range from which to request MIRA allocates subnets",
				Type:        schema.TypeString,
				Required:    true, // require field populated in terraform
			},
			"requestmask": {
				Description: "The CIDR of the range from which to request MIRA allocate subnets",
				Type:        schema.TypeString,
				Required:    true,
			},
			// these two resources are populated via api response from mira
			"message": {
				Description: "Mira Response Status Code (OK). Retrieved from MIRA API",
				Type:        schema.TypeString,
				Computed:    true, // get value from api
			},
			"payload": {
				Description: "Mira Available Subnets from within a specified range. Retrieved from MIRA API",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

// =========
// CRUD READ
// =========

func dataSourceMiraAvailableSubnetsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// ----------
	// GET CLIENT
	// ----------

	// use the meta value to retrieve your client from the mira provider configure method
	client := meta.(*miraclient.Client)

	// ------------------------
	// GET FIELDS FROM RESOURCE
	// ------------------------

	// get the ip range, from terraform resource, to request ip subnets from in mira
	requestRange := data.Get("requestrange").(string)
	requestMask  := data.Get("requestmask").(string)

	// --------------------------------
	// DO MIRA FREE SUBNETS API REQUEST
	// --------------------------------

	// add the mira range and netmask to the miraFreeSubnetsQuery datastructure
	miraFreeSubnetsQuery := &miraclient.RangeForAvailableMiraSubnetsQueryInput{
		RequestRange: requestRange,
		RequestMask:  requestMask,
	}

	// get free subnets from mira range, from api endpoint
	unmarshaledResponseData, err := client.GetAvailableSubnetsFromMiraRange(miraFreeSubnetsQuery)
	if err != nil {
		return diag.FromErr(err)
	}

	// -------------------------------
	// SET RESOURCE WITH API RESPONSES
	// -------------------------------

	// add the response message (a status string 'OK') to the terraform resource
	if err := data.Set("message", unmarshaledResponseData.Message); err != nil {
		return diag.FromErr(err)
	}

	// add the response payload (a list of available subnets) to the terraform resource
	if err := data.Set("payload", unmarshaledResponseData.Payload); err != nil {
		return diag.FromErr(err)
	}

	// ---------------------------------
	// SET ID TO UNIX TIME SO ALWAYS NEW
	// ---------------------------------

	// always run
	data.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	// ------------------------
	// RETURN INFO AND WARNINGS
	// ------------------------
	return diags
}

