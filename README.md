# Moesif AWS Lambda Middleware for Go

[![Built For][ico-built-for]][link-built-for]
[![Software License][ico-license]][link-license]
[![Source Code][ico-source]][link-source]

Go Middleware for AWS Lambda that automatically logs API calls 
and sends to [Moesif](https://www.moesif.com) for API analytics and log analysis. 

Designed for APIs that are hosted on AWS Lambda using Amazon API Gateway or Application Load Balancer
as a trigger.

## How to install

Run the following commands:
`moesif-aws-lambda-go` can be installed like any other Go library through go get:
```shell
go get github.com/moesif/moesif-aws-lambda-go
```

Or, if you are already using Go Modules, specify a version number as well:
```shell
go get github.com/moesif/moesif-aws-lambda-go@v1.0.2
```

## How to use

### 1. Add middleware to your Lambda application.

```go
package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
    "context"
    "strings"
    "github.com/aws/aws-lambda-go/events"
	moesifawslambda "github.com/moesif/moesif-aws-lambda-go"
)

func MoesifOptions() map[string]interface{} {
	var moesifOptions = map[string]interface{} {
		"Log_Body": true,
	}
	return moesifOptions
}

func HandleLambdaEvent(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       request.Body,
		StatusCode: 200,
		Headers: map[string] string {
			"Content-Type":   "application/json",
		},
	   }, nil
}

func main() {
	lambda.Start(moesifawslambda.MoesifLogger(HandleLambdaEvent, MoesifOptions()))
}
```

### 2. Set MOESIF_APPLICATION_ID environment variable 

Add a new environment variable with the name `MOESIF_APPLICATION_ID` and the value being your Moesif application id,
which can be found in the [_Moesif Portal_](https://www.moesif.com/).
After signing up for a Moesif account, your Moesif Application Id will be displayed during the onboarding steps. 

You can always find your Moesif Application Id at any time by logging 
into the [_Moesif Portal_](https://www.moesif.com/), click on the top right menu,
 and then clicking _Installation_.

## Optional: Capturing outgoing API calls
In addition to your own APIs, you can also start capturing calls out to third party services via the following method:

```go
func MoesifOptions() map[string]interface{} {
	var moesifOptions = map[string]interface{} {
		"Application_Id": "Your Moesif Application Id",
		"Log_Body": true,
	}
	return moesifOptions
}

moesifawslambda.StartCaptureOutgoing(MoesifOptions())
```

#### `moesifOption`
(__required__), _map[string]interface{}_, are the configuration options for your application. Please find the details below on how to configure options.

## Configuration options

Please note that the [request](https://github.com/aws/aws-lambda-go/blob/master/events/apigw.go#L6) and the [response](https://github.com/aws/aws-lambda-go/blob/master/events/apigw.go#L22) parameters in the configuration options are `aws-lambda-go` APIGatewayProxyRequest and APIGatewayProxyResponse respectively.

#### __`Should_Skip`__
(optional) _(request, response) => boolean_, a function that takes a request and a response,
and returns true if you want to skip this particular event.

#### __`Identify_User`__
(optional, but highly recommended) _(request, response) => string_, a function that takes a request and response, and returns a string that is the user id used by your system. While Moesif tries to identify users automatically, but different frameworks and your implementation might be very different, it would be helpful and much more accurate to provide this function.

#### __`Identify_Company`__
(optional) _(request, response) => string_, a function that takes a request and response, and returns a string that is the company id for this event.

#### __`Get_Metadata`__
(optional) _(request, response) => dictionary_, a function that takes a request and response, and
returns a dictionary (must be able to be encoded into JSON). This allows you
to associate this event with custom metadata. For example, you may want to save a VM instance_id, a trace_id, or a tenant_id with the request.

#### __`Get_Session_Token`__
(optional) _(request, response) => string_, a function that takes a request and response, and returns a string that is the session token for this event. Moesif tries to get the session token automatically, but if this doesn't work for your service, you should use this to identify sessions.

#### __`Mask_Event_Model`__
(optional) _(EventModel) => EventModel_, a function that takes an EventModel and returns an EventModel with desired data removed. The return value must be a valid EventModel required by Moesif data ingestion API. For details regarding EventModel please see the [Moesif Golang API Documentation](https://www.moesif.com/docs/api?go).

#### __`Debug`__
(optional) _boolean_, a flag to see debugging messages.

#### __`Log_Body`__
(optional) _boolean_, Default true. Set to false to remove logging request and response body to Moesif.

#### __`Capture_Outoing_Requests`__
(optional) _boolean_, Default False. Set to True to capture all outgoing API calls from your app to third parties like Stripe or to your own dependencies while using [net/http](https://golang.org/pkg/net/http/) package. The options below is applied to outgoing API calls.
When the request is outgoing, for options functions that take request and response as input arguments, the request and response objects passed in are [Request](https://golang.org/src/net/http/request.go) request and [Response](https://golang.org/src/net/http/response.go) response objects.

##### __`Should_Skip_Outgoing`__
(optional) _(request, response) => boolean_, a function that takes a request and response, and returns true if you want to skip this particular event.

##### __`Identify_User_Outgoing`__
(optional, but highly recommended) _(request, response) => string_, a function that takes request and response, and returns a string that is the user id used by your system. While Moesif tries to identify users automatically,
but different frameworks and your implementation might be very different, it would be helpful and much more accurate to provide this function.

##### __`Identify_Company_Outgoing`__
(optional) _(request, response) => string_, a function that takes request and response, and returns a string that is the company id for this event.

##### __`Get_Metadata_Outgoing`__
(optional) _(request, response) => dictionary_, a function that takes request and response, and
returns a dictionary (must be able to be encoded into JSON). This allows
to associate this event with custom metadata. For example, you may want to save a VM instance_id, a trace_id, or a tenant_id with the request.

##### __`Get_Session_Token_Outgoing`__
(optional) _(request, response) => string_, a function that takes request and response, and returns a string that is the session token for this event. Again, Moesif tries to get the session token automatically, but if you setup is very different from standard, this function will be very help for tying events together, and help you replay the events.

##### __`Mask_Event_Model_Outgoing`__
(optional) _(EventModel) => EventModel_, a function that takes an EventModel and returns an EventModel with desired data removed. The return value must be a valid EventModel required by Moesif data ingestion API. For details regarding EventModel please see the [Moesif Golang API Documentation](https://www.moesif.com/docs/api?go).

##### __`Log_Body_Outgoing`__
(optional) _boolean_, Default true. Set to false to remove logging request and response body to Moesif.

## Update User

### UpdateUser method
Create or update a user profile in Moesif.
The metadata field can be any customer demographic or other info you want to store.
Only the `UserId` field is required.
This method is a convenient helper that calls the Moesif API lib.
For details, visit the [Go API Reference](https://www.moesif.com/docs/api?go#update-a-user).

```go
import (
	moesifawslambda "github.com/moesif/moesif-aws-lambda-go"
)

func literalFieldValue(value string) *string {
    return &value
}

var moesifOptions = map[string]interface{} {
	"Application_Id": "Your Moesif Application Id",
}

// Campaign object is optional, but useful if you want to track ROI of acquisition channels
// See https://www.moesif.com/docs/api#users for campaign schema
campaign := models.CampaignModel {
  UtmSource: literalFieldValue("google"),
  UtmMedium: literalFieldValue("cpc"), 
  UtmCampaign: literalFieldValue("adwords"),
  UtmTerm: literalFieldValue("api+tooling"),
  UtmContent: literalFieldValue("landing"),
}
  
// metadata can be any custom dictionary
metadata := map[string]interface{}{
  "email": "john@acmeinc.com",
  "first_name": "John",
  "last_name": "Doe",
  "title": "Software Engineer",
  "sales_info": map[string]interface{}{
      "stage": "Customer",
      "lifetime_value": 24000,
      "account_owner": "mary@contoso.com",
  },
}

// Only UserId is required
user := models.UserModel{
  UserId:  "12345",
  CompanyId:  literalFieldValue("67890"), // If set, associate user with a company object
  Campaign:  &campaign,
  Metadata:  &metadata,
}

// Update User
moesifawslambda.UpdateUser(&user, moesifOption)
```

### UpdateUsersBatch method
Similar to UpdateUser, but used to update a list of users in one batch. 
Only the `UserId` field is required.
This method is a convenient helper that calls the Moesif API lib.
For details, visit the [Go API Reference](https://www.moesif.com/docs/api?go#update-users-in-batch).

```go

import (
	moesifawslambda "github.com/moesif/moesif-aws-lambda-go"
)

func literalFieldValue(value string) *string {
    return &value
}

var moesifOptions = map[string]interface{} {
	"Application_Id": "Your Moesif Application Id",
}

// List of Users
var users []*models.UserModel

// Campaign object is optional, but useful if you want to track ROI of acquisition channels
// See https://www.moesif.com/docs/api#users for campaign schema
campaign := models.CampaignModel {
  UtmSource: literalFieldValue("google"),
  UtmMedium: literalFieldValue("cpc"), 
  UtmCampaign: literalFieldValue("adwords"),
  UtmTerm: literalFieldValue("api+tooling"),
  UtmContent: literalFieldValue("landing"),
}
  
// metadata can be any custom dictionary
metadata := map[string]interface{}{
  "email": "john@acmeinc.com",
  "first_name": "John",
  "last_name": "Doe",
  "title": "Software Engineer",
  "sales_info": map[string]interface{}{
      "stage": "Customer",
      "lifetime_value": 24000,
      "account_owner": "mary@contoso.com",
  },
}

// Only UserId is required
userA := models.UserModel{
  UserId:  "12345",
  CompanyId:  literalFieldValue("67890"), // If set, associate user with a company object
  Campaign:  &campaign,
  Metadata:  &metadata,
}

users = append(users, &userA)

// Update User
moesifawslambda.UpdateUsersBatch(users, moesifOption)
```

## Update Company

### UpdateCompany method
Create or update a company profile in Moesif.
The metadata field can be any company demographic or other info you want to store.
Only the `CompanyId` field is required.
This method is a convenient helper that calls the Moesif API lib.
For details, visit the [Go API Reference](https://www.moesif.com/docs/api?go#update-a-company).

```go
import (
	moesifawslambda "github.com/moesif/moesif-aws-lambda-go"
)

func literalFieldValue(value string) *string {
    return &value
}

var moesifOptions = map[string]interface{} {
	"Application_Id": "Your Moesif Application Id",
}

// Campaign object is optional, but useful if you want to track ROI of acquisition channels
// See https://www.moesif.com/docs/api#update-a-company for campaign schema
campaign := models.CampaignModel {
  UtmSource: literalFieldValue("google"),
  UtmMedium: literalFieldValue("cpc"), 
  UtmCampaign: literalFieldValue("adwords"),
  UtmTerm: literalFieldValue("api+tooling"),
  UtmContent: literalFieldValue("landing"),
}
  
// metadata can be any custom dictionary
metadata := map[string]interface{}{
  "org_name": "Acme, Inc",
  "plan_name": "Free",
  "deal_stage": "Lead",
  "mrr": 24000,
  "demographics": map[string]interface{}{
      "alexa_ranking": 500000,
      "employee_count": 47,
  },
}

// Prepare company model
company := models.CompanyModel{
	CompanyId:		  "67890",	// The only required field is your company id
	CompanyDomain:  literalFieldValue("acmeinc.com"), // If domain is set, Moesif will enrich your profiles with publicly available info 
	Campaign: 		  &campaign,
	Metadata:		    &metadata,
}

// Update Company
moesifawslambda.UpdateCompany(&company, moesifOption)
```

### UpdateCompaniesBatch method
Similar to UpdateCompany, but used to update a list of companies in one batch. 
Only the `CompanyId` field is required.
This method is a convenient helper that calls the Moesif API lib.
For details, visit the [Go API Reference](https://www.moesif.com/docs/api?go#update-companies-in-batch).

```go

import (
	moesifawslambda "github.com/moesif/moesif-aws-lambda-go"
)

func literalFieldValue(value string) *string {
    return &value
}

var moesifOptions = map[string]interface{} {
	"Application_Id": "Your Moesif Application Id",
}

// List of Companies
var companies []*models.CompanyModel

// Campaign object is optional, but useful if you want to track ROI of acquisition channels
// See https://www.moesif.com/docs/api#update-a-company for campaign schema
campaign := models.CampaignModel {
  UtmSource: literalFieldValue("google"),
  UtmMedium: literalFieldValue("cpc"), 
  UtmCampaign: literalFieldValue("adwords"),
  UtmTerm: literalFieldValue("api+tooling"),
  UtmContent: literalFieldValue("landing"),
}
  
// metadata can be any custom dictionary
metadata := map[string]interface{}{
  "org_name": "Acme, Inc",
  "plan_name": "Free",
  "deal_stage": "Lead",
  "mrr": 24000,
  "demographics": map[string]interface{}{
      "alexa_ranking": 500000,
      "employee_count": 47,
  },
}

// Prepare company model
companyA := models.CompanyModel{
	CompanyId:		  "67890",	// The only required field is your company id
	CompanyDomain:  literalFieldValue("acmeinc.com"), // If domain is set, Moesif will enrich your profiles with publicly available info 
	Campaign: 		  &campaign,
	Metadata:		    &metadata,
}

companies = append(companies, &companyA)

// Update Companies
moesifawslambda.UpdateCompaniesBatch(companies, moesifOption)
```

## Examples

- [A complete example is available on GitHub](https://github.com/Moesif/moesif-aws-lambda-go-example).

## Other integrations

To view more documentation on integration options, please visit __[the Integration Options Documentation](https://www.moesif.com/docs/getting-started/integration-options/).__

[ico-built-for]: https://img.shields.io/badge/built%20for-aws%20lambda-blue.svg
[ico-license]: https://img.shields.io/badge/License-Apache%202.0-green.svg
[ico-source]: https://img.shields.io/github/last-commit/moesif/moesif-aws-lambda-go.svg?style=social

[link-built-for]: https://aws.amazon.com/lambda/
[link-license]: https://raw.githubusercontent.com/Moesif/moesif-aws-lambda-go/master/LICENSE
[link-source]: https://github.com/moesif/moesif-aws-lambda-go