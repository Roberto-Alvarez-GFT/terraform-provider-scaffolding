package mira

import (
	"context"
	// "errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	// the db mira client
	"terraform-provider-mira/miraclient"
)

// ============================================================
// FUNCTION ASSIGNED TO RESOURCE IN provider.go [RETURN SCHEMA]
// ============================================================

func resourceMiraAllocatedSubnet() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "A resource in the Terraform provider Mira for dynamically assigning a subnet from with a specified address range.",

		// function names in this file that are assigned to CRUD calls 
		CreateContext: resourceMiraAllocatedSubnetCreate,
		ReadContext:   resourceMiraAllocatedSubnetRead,
		UpdateContext: resourceMiraAllocatedSubnetUpdate,
		DeleteContext: resourceMiraAllocatedSubnetDelete,

		// the resources schema map of its fields
		Schema: map[string]*schema.Schema{
			// populate these fields from text in terraform hcl [resources]
			"addressid": {
				Type:         schema.TypeString,
				Required:     true, // require fields are populated in terraform
				Description: "A 7 digit integer that identifies the physical location in the world (a site id for an address)",
			},
			"comment": {
				Type:         schema.TypeString,
				Required:     true,
				Description: "A description for the subnet use",
			},
			"requestrange": {
				Type:         schema.TypeString,
				Required:     true,
				Description: "!!IMPORTANT!! Mira Range from which to assign a subnet",
			},
			"requestmask": {
				Type:         schema.TypeString,
				Required:     true,
				Description: "!!IMPORTANT!! Subnet mask for Mira Range from which to assign a subnet",
			},
			"subnetname": {
				Type:         schema.TypeString,
				Required:     true,
				Description: "A description matching the comment field",
			},
			"template": {
				Type:         schema.TypeString,
				Required:     true,
				Description: "One of the values: U25_DEV_GCP U25_UAT_GCP U25_PRD_GCP which are required preconfigured template names",
			},
			// IMPORTANT: following values are populated by the api calls to MIRA
			"miraassignedsubnet": {
				Type:         schema.TypeString,
				Computed:     true, // fields are populated via api response from mira
				Description: "A subnet from within the requestrange, assigned by mira to this projects network",
			},
			"miraassignedsubnetmask": {
				Type:         schema.TypeString,
				Computed:     true,
				Description: "A subnetmask, assigned by mira to this projects network",
			},

			// Commented out: not required for the functionality used, eg: ip and nm octets are split by the client; the
			//                two miraassigned* resources above hold the ip address string (thats been checked by the client)

			// "alsoQip": {
			// 	Type:         schema.TypeBool,
			// 	Required:     true,
			// 	Description: "Always False",
			// },
			// "building": {
			// 	Type:         schema.TypeString,
			// 	Required:     true,
			// 	Description: "Not applicable - set this to the empty string - this is a place holder variable that the api requires",
			// },
			// "dhcp": {
			// 	Type:         schema.TypeBool,
			// 	Required:     true,
			// 	Description: "Always false",
			// },
			// "dhcpServer": {
			// 	Type:         schema.TypeString,
			// 	Required:     true,
			// 	Description: "Always empty string",
			// },
			// "dhcpTemplate": {
			// 	Type:         schema.TypeString,
			// 	Required:     true,
			// 	Description: "Always empty string",
			// },
			// "floor": {
			// 	Type:         schema.TypeString,
			// 	Required:     true,
			// 	Description: "Always empty string",
			// },
			// "ip1": {
			// 	Type:         schema.TypeString,
			// 	Required:     true,
			// 	Description: "First octect of IP as string",
			// },
			// "ip2": {
			// 	Type:         schema.TypeString,
			// 	Required:     true,
			// 	Description: "Second octect of IP as string",
			// },
			// "ip3": {
			// 	Type:         schema.TypeString,
			// 	Required:     true,
			// 	Description: "Third octect of IP as string",
			// },
			// "ip4": {
			// 	Type:         schema.TypeString,
			// 	Required:     true,
			// 	Description: "Fourth octect of IP as string",
			// },
			// "ipAddressSchema": {
			// 	Type:         schema.TypeInt,
			// 	Required:     true,
			// 	Description: "Always number 2 schema required",
			// },
			// "netmask1": {
			// 	Type:         schema.TypeString,
			// 	Required:     true,
			// 	Description: "First octect of Netmask as string",
			// },
			// "netmask2": {
			// 	Type:         schema.TypeString,
			// 	Required:     true,
			// 	Description: "Second octect of Netmask as string",
			// },
			// "netmask3": {
			// 	Type:         schema.TypeString,
			// 	Required:     true,
			// 	Description: "Third octect of Netmask as string",
			// },
			// "netmask4": {
			// 	Type:         schema.TypeString,
			// 	Required:     true,
			// 	Description: "Fourth octect of Netmask as string",
			// },
			// "recordId": {
			// 	Type:         schema.TypeString,
			// 	Required:     true,
			// 	Description: "Empty string",
			// },
			// "room": {
			// 	Type:         schema.TypeString,
			// 	Required:     true,
			// 	Description: "Empty string",
			// },
			// "subnetClass": {
			// 	Type:         schema.TypeInt,
			// 	Required:     true,
			// 	Description: "Always intiger 38 (which is GCP)",
			// },
			// "subnetNameChanged": {
			// 	Type:         schema.TypeBool,
			// 	Required:     true,
			// 	Description: "Always false",
			// },
			// "vlan": {
			// 	Type:         schema.TypeString,
			// 	Required:     true,
			// 	Description: "Empty string",
			// },
		},
		UseJSONNumber: true,
	}
}

// ===========
// CRUD CREATE
// ===========

func resourceMiraAllocatedSubnetCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {

	// Warning or info can be collected in a slice type (only returned if no error)
	var diags diag.Diagnostics

	// ----------
	// GET CLIENT
	// ----------

	// use the meta value to retrieve your client from the mira provider configure method
	// client := miraclient.NewClient(apiClient.Username, apiClient.Password, apiClient.Uri)
	client := meta.(*miraclient.Client)

	// ------------------------
	// GET FIELDS FROM RESOURCE
	// ------------------------

	requestRange	 := data.Get("requestrange").(string)
	requestMask	 := data.Get("requestmask").(string)
	addressID	 := data.Get("addressid").(string)
	comment		 := data.Get("comment").(string)
	subnetName       := data.Get("subnetname").(string)
	template         := data.Get("template").(string)

	// write logs using the tflog package
	// see https://pkg.go.dev/github.com/hashicorp/terraform-plugin-log/tflog
	// for more information
	// tflog.Trace(ctx, "created a resource")

	// -------------------------------------------------
	// DO MIRA CREATE SUBNET ASSIGNMENT API POST REQUEST
	// -------------------------------------------------

	// add the mira range and netmask and other details to the mira assignment datastructure
	miraAssignSubnetRequestInput := &miraclient.MiraSubnetAssignmentPostInput{
		RequestRange: requestRange,
		RequestMask:  requestMask,
		AddressID:    addressID,
		Comment:      comment,
		SubnetName:   subnetName,
		Template:     template,
	}

	// assign a free subnet from within the RequestRange to this project, to do this we hit the 
	// get free subnets from mira range api endpoint and then choose the first available subnet
	// and submit this to mira via a post request to create the assignment, return subnet/error
	chosenSubnet, err := client.CreateMiraSubnetAssignment(miraAssignSubnetRequestInput)
	if err != nil {
		return diag.FromErr(err)
	}

	// IMPORTANT: I am setting the subnet mask here because i dont know where it comes from currently
	//            to remedy this i will be speaking to John
	chosenSubnetMask := "255.255.255.224"

	// ---------------------------------------------
	// SET TERRAFORM RESOURCE DATA WITH API RESPONSE
	// ---------------------------------------------

	// add the chosen subnet from the api response to the resource field
	if err := data.Set("miraassignedsubnet", chosenSubnet); err != nil {
		return diag.FromErr(err)
	}

	// add the subnet mask of the chosen subnet to the resource field
	if err := data.Set("miraassignedsubnetmask", chosenSubnetMask); err != nil {
		return diag.FromErr(err)
	}

	// --------------------------------------
	// SET RESOURCE ID TO "SUBNET-SUBNETMASK"
	// --------------------------------------

	// run when all conditions are met
	data.SetId(chosenSubnet + "-" + chosenSubnetMask)

	// ---------------------------------------------------
	// RETURN DIAGS FOR INFO AND WARNINGS (no errors seen)
	// ---------------------------------------------------

	return diags
}

// =========
// CRUD READ
// =========

func resourceMiraAllocatedSubnetRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// --------------
	// GET THE CLIENT
	// --------------

	// use the meta value to retrieve your client from the mira provider configure method
//	client := meta.(*miraclient.Client)

	// --------------------------------
	// GET THE FIELDS FROM THE RESOURCE
	// --------------------------------

//	SubnetAddress	 := data.Get("miraassignedsubnet").(string)
	// SubnetMask	 := data.Get("miraassignedsubnetmask").(string)  // not used

	// ---------------------------------------
	// DO THE API REQUEST TO GET SUBNET RECORD
	// ---------------------------------------

	// add subnet address to query datastructure
//	findMiraSubnetByIpQueryInput := &miraclient.GetMiraSubnetFromIPAddressQueryInput{
//		IpAddress: SubnetAddress,
//	}

	// do the api request to get the subnet record for an ip from MIRA
//	returnedSubnet, err := client.GetMiraSubnetRecordFromIPAddress(findMiraSubnetByIpQueryInput)
//	if err != nil {
//		// return diag.FromErr(err)
//	}

	// ---------------
	// DO SIMPLE CHECK
	// ---------------

	// check returned address matches submitted (terraform should probably do this)
	// if returnedSubnet.IpAddress != SubnetAddress {
		// return diag.FromErr(errors.New("Returned Subnet did not match stored subnet"))
	// }

	// ---------------------------------------------
	// SET TERRAFORM RESOURCE DATA WITH API RESPONSE
	// ---------------------------------------------

	requestRange	 := data.Get("requestrange").(string)
	requestMask	 := data.Get("requestmask").(string)
	// add the returned subnet IpAddress to the resource (to be compaired, and evaluate to no change... i think)
	if err := data.Set("miraassignedsubnet", requestRange); err != nil {
		// return diag.FromErr(err)
	}

	// add the returned subnet mask to the resource to be compaired to previous value
	if err := data.Set("miraassignedsubnetmask", requestMask); err != nil {
		// return diag.FromErr(err)
	}

	// --------------------------------------
	// SET RESOURCE ID TO "SUBNET-SUBNETMASK"
	// --------------------------------------

	// all conditions met, set the resource id
	data.SetId(requestRange + "-" + requestMask)

	// -----------------------------------------
	// RETURN INFO AND WARNINGS (no errors seen)
	// -----------------------------------------

	return diags
}

func resourceMiraAllocatedSubnetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the mira provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented, you must contact the CNE Team to change an allocation")
}

func resourceMiraAllocatedSubnetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the mira provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented, you must contact the CNE team to remove your allocation")
}
