package miraclient

import (
	"fmt"
	"errors"
	"os"
	"net"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"time"
	"bytes"
	"strings"
)

const baseURL   string = "http://10.156.0.3/"
const userAgent string = "terraform-provider-mira"

// **************************
// CREATE A NEW CLIENT STRUCT
// **************************

// client configuration for connection setup
type Client struct {
	Username   string
	Password   string
	UserAgent  string
	URL        string
	HTTPClient *http.Client
}

// =========================================
// CREATE A NEW CLIENT (and populate struct)
// =========================================

// Create a new client, and load username, password, and userAgent from environment variables
func NewClient() (*Client, error) {

	// get mira user and pw from environment
	username  := os.Getenv("MIRA_USERNAME")
	password  := os.Getenv("MIRA_PASSWORD")
	userAgent := os.Getenv("TERRAFORM_USERAGENT_MIRA")

	// if env does not have username or password return error
	if (username == "") || (password == "") {
		return nil, errors.New("no username or password in env")
	}

	// return error if no useragent is provided in the environment (should be: terraform-provider-mira)
	if userAgent == "" {
		return nil, errors.New("no useragent string in env")
	}

	// create client
	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		URL: baseURL,
		UserAgent: userAgent,
		Username:  username,
		Password:  password,
	}

	// return a pointer to the client
	return &c, nil
}

// ------------------------------------------------
// IP CHECKER FUNCTION FOR USE IN ALL METHODS BELOW
// ------------------------------------------------

// func to test if a string is an IP address (used in create assignment)
func checkIPAddress(ip string) bool {
	if net.ParseIP(ip) == nil {
		return false
	} else {
		return true
	}
}

// ======================================================
// METHOD: doRequest [DO HTTP REQUEST, RETURN BODY BYTES]
// ======================================================

// Add method to the Client struct, that executes 
// the http reqest and returns the body bytes
func (c *Client) doRequest(req *http.Request) ([]byte, error) {

	// use the http client to 'do' the request
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// read the body of the reponse into a variable
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// return error if status was not http:200
	if res.StatusCode != http.StatusOK {
		return  nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	// give body of response to calling function
	return body, nil
}

// *********************************************************************
// CREATE INPUT AND OUTPUT STRUCTS FOR: GetAvailableSubnetsFromMiraRange
// *********************************************************************

// CIDR range from which to request a list of available subnets
type RangeForAvailableMiraSubnetsQueryInput struct {
	RequestRange string
	RequestMask  string
}

// A list of subnets from MIRA that are available for use 
type AvailableSubnetsResponseFromMira struct {
	Message      string   `json:"message"`
	Payload      []string `json:"payload"`
}

// =================================================================================================
// METHOD: GetAvailableSubnetsFromMiraRange [REQUEST FREE SUBNETS BY RANGE, RETURN FREE SUBNET LIST]
// =================================================================================================

// Create a http request, add authentication details and ranges to request a free subnet from
func (c *Client) GetAvailableSubnetsFromMiraRange(miraRange *RangeForAvailableMiraSubnetsQueryInput) (*AvailableSubnetsResponseFromMira, error) {

	// -----------
	// CHECK INPUT
	// -----------

	// check that the mira range and mask are in ip address format
	if !(checkIPAddress(miraRange.RequestRange)) {
		return nil, fmt.Errorf("Error: %s is not in IP address format, in GetAvailableSubnetsFromMiraRange RangeRequest", miraRange.RequestRange)
	}

	// check that the mira range and mask are in ip address format
	if !(checkIPAddress(miraRange.RequestMask)) {
		return nil, fmt.Errorf("Error: %s is not in IP address format, in GetAvailableSubnetsFromMiraRange RangeMask", miraRange.RequestMask)
	}

	// ------------------------
	// CREATE HTTP REQUEST DATA
	// ------------------------

	// create MIRA search free subnet query string, using the mira supernamt and cidr
	// url := fmt.Sprintf(baseURL+"searchFreeSubnet?range=%s&netmaskNew=%s", miraRange.RequestRange, miraRange.RangeMask)
	url    := fmt.Sprintf(baseURL)
	method := "GET"

	// create a new get request object for the url above
	freeSubnetReq, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	// ----------------
	// SET AUTH HEADERS
	// ----------------

	// set auth header (not needed so commented)
	freeSubnetReq.SetBasicAuth(c.Username, c.Password)
	// set content to json
	freeSubnetReq.Header.Set("Content-Type", "application/json")

	// --------------
	// DO API REQUEST
	// --------------

	// do http request (func above this one) and return a string of the body text
	freeSubnetRespBody, err := c.doRequest(freeSubnetReq)
	if err != nil {
		return nil, err
	}

	// --------------------------
	// UNMARSHAL JSON FROM BYTES
	// --------------------------

	// create a struct for the api response
	var unmarshaledResponseData AvailableSubnetsResponseFromMira

	// unmarshal the data from the response body into responseData
	err = json.Unmarshal(freeSubnetRespBody, &unmarshaledResponseData)
	if err != nil {
		return nil, err
	}

	// -------------------
	// CHECK RESPONSE DATA
	// -------------------

	// get the responses "message" field and check its a status string 'OK'
	statusOK := unmarshaledResponseData.Message
	// add the response payload (a list of available subnets) to the terraform resource
	freeSubnetsList := unmarshaledResponseData.Payload

	// check the api response was 'OK' 
	if (statusOK != "OK") {
		return nil, fmt.Errorf("Error: mira api status was not 'OK' status: %s", statusOK)
	}
	// itterate over response body and ...
	for index, element := range freeSubnetsList {
		// check that the element is actaully an IP
		if !(checkIPAddress(element)) {
			return nil, fmt.Errorf("Error: %s is not Subnet, at position %d mira free subnets payload %d", element, index, freeSubnetsList)
		}
	}

	// --------------------
	// RETURN GOOD RESPONSE
	// --------------------

	// return that shiz
	return &unmarshaledResponseData, nil
}

// *****************************************************************
// CREATE INPUT AND POSTDATA STRUCTS FOR: CreateMiraSubnetAssignment
// *****************************************************************

// struct for mira subnet assignment post input variables
type MiraSubnetAssignmentPostInput struct {
	RequestRange      string
	RequestMask       string
	AddressID         string
	Comment           string
	SubnetName        string
	Template          string
}

// input from MiraSubnetAssignmentPostInput and static values 
// will be added to this post data to fulfil the required fields
// eg: ip and netmask addresses will be split into octets
type MiraSubnetAssignmentPostData struct {
	AddressID         string `json:"addressID"`
	AlsoQip           bool   `json:"alsoQip"`
	Building          string `json:"building"`
	Comments          string `json:"comments"`
	Dhcp              bool   `json:"dhcp"`
	DhcpServer        string `json:"dhcpServer"`
	DhcpTemplate      string `json:"dhcpTemplate"`
	Floor             string `json:"floor"`
	Ip1               string `json:"ip1"`
	Ip2               string `json:"ip2"`
	Ip3               string `json:"ip3"`
	Ip4               string `json:"ip4"`
	IpAddressSchema   int    `json:"ipAddressSchema"`
	Netmask1          string `json:"netmask1"`
	Netmask2          string `json:"netmask2"`
	Netmask3          string `json:"netmask3"`
	Netmask4          string `json:"netmask4"`
	Range             string `json:"range"`
	RecordId          string `json:"recordId"`
	Room              string `json:"room"`
	SubnetClass       int    `json:"subnetClass"`
	SubnetName        string `json:"subnetName"`
	SubnetNameChanged bool   `json:"subnetNameChanged"`
	Template          string `json:"template"`
	Vlan              string `json:"vlan"`
}

// ===================================================================================
// METHOD: CreateMiraSubnetAssignment [REQUEST SUBNETS ASSIGNMENT, RETURN HTTP STATUS]
// ===================================================================================

func (c *Client) CreateMiraSubnetAssignment(postInput *MiraSubnetAssignmentPostInput) (string, error) {

	// -----------------------------------------------------
	// PUT INPUT STRUCT INTO INDIVIDUAL VARS FOR READABILITY
	// -----------------------------------------------------

	// get input data required for api query from terrafrom resource (provided by module)
	mirarange  := postInput.RequestRange
	rangemask  := postInput.RequestMask
	addressID  := postInput.AddressID
	comment	   := postInput.Comment
	subnetname := postInput.SubnetName
	template   := postInput.Template


	// ---------------
	// CHECK INPUT IPS
	// ---------------

	// check that the range to be supplied to mira is in ip address format
	if !(checkIPAddress(mirarange)) {
		return "", fmt.Errorf("Error: %s is not valid mira range", mirarange)
	}

	// check that the range mask to be supplied to mira is in ip address format
	if !(checkIPAddress(rangemask)) {
		return "", fmt.Errorf("Error: %s is not a valid mira mask", rangemask)
	}

	// -------------------------------------------
	// DO MIRA FREE SUBNETS FROM RANGE API REQUEST
	// -------------------------------------------

	// add inpput vars to query input struct
	rangeForAvailableSubnets := RangeForAvailableMiraSubnetsQueryInput{
		RequestRange: mirarange,
		RequestMask: rangemask,
	}

	// get free subnets from mira range, from api endpoint
	unmarshaledResponseData, err := c.GetAvailableSubnetsFromMiraRange(&rangeForAvailableSubnets)
	if err != nil {
		return "", err
	}

	// --------------------------------------------
	// CHECK IF THE RESPONSE DATA CONTAINS A SUBNET
	// --------------------------------------------

	// add the response payload (a list of available subnets) to the terraform resource
	freeSubnetsList := unmarshaledResponseData.Payload

	// check the api response contained a list of available subnets
	if (len(freeSubnetsList) == 0) {
		return "", fmt.Errorf("Error: mira api returned an empty subnet array: [ %s ]", freeSubnetsList)
	}

	// ===========================================================================
	// FOR LOOP TO HERE FROM DO API - repeat using next range if subnet list is 0
	// ===========================================================================

	// ------------------------------------------
	// CHOOSE SUBNET FROM FREE SUBNETS API OUTPUT
	// ------------------------------------------

	// choose the first available subnet in the list
	chosenSubnet := freeSubnetsList[0]

	// check that the chosen IP is actaully an IP... just for good measure
	if !(checkIPAddress(chosenSubnet)) {
		return "", fmt.Errorf("Error: %s is not Subnet, but was about to be submitted to mira", chosenSubnet)
	}

	// -------------------------
	// PREPARE POST REQUEST DATA
	// -------------------------

	// split ipv4 subnet address and mask into individual strings per octet 
	var ipoctets []string = strings.Split(chosenSubnet, ".")
	var nmoctets []string = strings.Split(rangemask, ".")

	// create MIRA assign subnet post url string
	// url := fmt.Sprintf(baseURL+"searchFreeSubnet?range=%s&netmaskNew=%s", miraRange.RequestRange, miraRange.RangeMask)
	url    := fmt.Sprintf(baseURL)
	method := "POST"

	// Encode the data for the post, from a struct to json
	postBody, err := json.Marshal(MiraSubnetAssignmentPostData{
		AddressID: addressID,		// 7 digit ID for physical location, can be prepopulated: eg: all locations eu-region3 get "765431"
		AlsoQip: false,			// always false
		Building: "",			// always empty
		Comments: comment,		// comment field from resource, populated with the subnets purpose
		Dhcp: false,			// always false
		DhcpServer: "",			// always empty
		DhcpTemplate: "",		// always empty
		Floor: "",			// always empty
		Ip1: ipoctets[0],		// first octect
		Ip2: ipoctets[1],		// second octet
		Ip3: ipoctets[2],		// third octet
		Ip4: ipoctets[3],		// fourth octet
		IpAddressSchema: 2,		// always use schema 2
		Netmask1: nmoctets[0],		// first octect
		Netmask2: nmoctets[1],		// second octet
		Netmask3: nmoctets[2],		// third octet
		Netmask4: nmoctets[3],		// fourth octet
		Range: mirarange,		// the range to request a subnet from, provided by default map in var in terraform module
		RecordId: "",			// always empty
		Room: "",			// always empty
		SubnetClass: 38,		// always intiger 38 (which is GCP)
		SubnetName: subnetname,		// a description matching the comment field
		SubnetNameChanged: false,	// always false
		Template: template,		// one of the values: U25_DEV_GCP U25_UAT_GCP U25_PRD_GCP which are required preconfigured template names
		Vlan: "",			// always empty
	})
	// check post marshaled to bytes ok
	if err != nil {
		return "", err
	}

	// create a new post request object for the url and method above
	assignSubnetReq, err := http.NewRequest(method, url, bytes.NewBuffer(postBody))
	if err != nil {
		return "", err
	}

	// ------------------------------
	// SET AUTH AND TYPE JSON HEADERS
	// ------------------------------

	// set auth header 
	assignSubnetReq.SetBasicAuth(c.Username, c.Password)
	// set content to json
	assignSubnetReq.Header.Set("Content-Type", "application/json")

	// --------------------------------
	// DO POST REQUEST TO ASSIGN SUBNET
	// --------------------------------

	// do http post request to assign the subnet
	_, err = c.doRequest(assignSubnetReq)
	if err != nil {
		// return "", err // NEVER USE THIS IN PRODUCTION WHILE COMMENTED - ONLY TEST
	}

	// IMPORTANT if there was no error the subnet is now assigned in mira but not in terraform
	return chosenSubnet, nil
}

// *********************************************************************
// CREATE INPUT AND OUTPUT STRUCTS FOR: GetMiraSubnetRecordFromIPAddress
// *********************************************************************

// the subnet address, submitted as just an ip
type GetMiraSubnetFromIPAddressQueryInput struct {
	IpAddress  string `json:"ipaddress"`
}

// the record from mira for the subnet from whence the ip came
type MiraSubnetFoundByIPAddressResponseData struct {
	IpAddress	string `json:"address"`
	IpMask		string `json:"mask"`
	Country		string `json:"country"`
	Description	string `json:"description"`
	Legacy		string `json:"legacy"`
	Layout		string `json:"layout"`
	RecordId	int    `json:"recordId"`
	SecurityDomain	string `json:"securityDomain"`
	SecurityZone	string `json:"securityZone"`
	SubnetClass	string `json:"subnetClass"`
	Tenant		string `json:"tenant"`
	QipInstance	string `json:"qipInstance"`
}

// ========================================================================================================
// METHOD: GetMiraSubnetRecordFromIPAddress [REQUEST SUBNETS RECORD FOR IP, RETURN SUBNET RECORD FROM MIRA]
// ========================================================================================================

// Create a http request, add authentication details and ranges to request a free subnet from
func (c *Client) GetMiraSubnetRecordFromIPAddress(queryInput *GetMiraSubnetFromIPAddressQueryInput) (*MiraSubnetFoundByIPAddressResponseData, error) {

	// -------------------
	// CREATE HTTP REQUEST
	// -------------------

	// create MIRA search subnet by address string, using the subnet address and cidr mask
	url := fmt.Sprintf(baseURL+ "search?containsIP=", queryInput.IpAddress)
	method := "GET"

	// create a new get request object for the url above
	getSubnetByIpReq, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	// ----------------
	// SET AUTH HEADERS
	// ----------------

	// set auth header (not needed so commented)
	getSubnetByIpReq.SetBasicAuth(c.Username, c.Password)
	// set content to json
	getSubnetByIpReq.Header.Set("Content-Type", "application/json")

	// -------------------
	// DO MIRA API REQUEST
	// -------------------

	// do http request and return a string of the body text
	getSubnetByIpReqBody, err := c.doRequest(getSubnetByIpReq)
	if err != nil {
		return nil, err
	}

	// --------------------
	// UNMARSHAL JSON BYTES
	// --------------------

	// create a struct for the api response body bytes
	var unmarshaledResponseData MiraSubnetFoundByIPAddressResponseData

	// unmarshal the data from the response body json bytes into struct
	err = json.Unmarshal(getSubnetByIpReqBody, &unmarshaledResponseData)
	if err != nil {
		return nil, err
	}

	// ----------------
	// RETURN BODY DATA
	// ----------------

	// return that shiz if no errors
	return &unmarshaledResponseData, nil
}

