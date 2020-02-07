#### Moesif AWS Lambda Middleware

[![Built For][ico-built-for]][link-built-for]
[![Software License][ico-license]][link-license]
[![Source Code][ico-source]][link-source]

Middleware (Go) to automatically log API calls from AWS Lambda functions
and sends to [Moesif](https://www.moesif.com) for API analytics and log analysis. 

Designed for APIs that are hosted on AWS Lambda using Amazon API Gateway as a trigger.

This middleware expects the
[Lambda proxy integration type.](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-set-up-simple-proxy.html#api-gateway-set-up-lambda-proxy-integration-on-proxy-resource)
If you're using AWS Lambda with API Gateway, you are most likely using the proxy integration type.

## How to install

```shell
go get github.com/moesif/moesif-aws-lambda-go
```

## How to use

The following shows how to import the module and use:

### 1. Import the module:

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

// Mask Event Model
func maskEventModel(eventModel models.EventModel) models.EventModel {
	return eventModel
}

// Set User Id
func identifyUser(request events.APIGatewayProxyRequest, response events.APIGatewayProxyResponse) string{
	return "golangapiuser"
}

// Set Company Id
func identifyCompany(request events.APIGatewayProxyRequest, response events.APIGatewayProxyResponse) string{
	return "golangapicompany"
}

// Set Session Token
func getSessionToken(request events.APIGatewayProxyRequest, response events.APIGatewayProxyResponse) string{
	return "XXXXXXXXXXXXXXXX"
}

// Skip Event
func shouldSkip(request events.APIGatewayProxyRequest, response events.APIGatewayProxyResponse) bool{
	return strings.Contains(request.Path, "skip")
}

// Set Metadata
func getMetadata(request events.APIGatewayProxyRequest, response events.APIGatewayProxyResponse) map[string]interface{} {
	
	var innerNestedFields = map[string] interface{} {
		"nestedInner": "test",
	}

	var nestedFields = map[string] interface{} {
		"inner":  innerNestedFields,
	}
	
	var metadata = map[string]interface{} {
		"foo" : "bar",
		"user": "golangapiuser",
		"test": nestedFields,
	}
	return metadata
}

func MoesifOptions() map[string]interface{} {
	var moesifOptions = map[string]interface{} {
		"Api_Version": "1.0.0",
		"Get_Metadata": getMetadata,
		"Should_Skip": shouldSkip,
		"Identify_User": identifyUser,
		"Identify_Company": identifyCompany,
		"Get_Session_Token": getSessionToken,
		"Mask_Event_Model": maskEventModel,
		"Debug": true,
		"Log_Body": true,
	}
	return moesifOptions
}

func HandleLambdaEvent(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       request.Body,
		StatusCode: 200,
		Headers: map[string] string {
			"RspHeader1":     "RspHeaderValue1",
			"Content-Type":   "application/json",
			"Content-Length": "1000",
		},
	   }, nil
}

func main() {
	lambda.Start(moesifawslambda.MoesifLogger(HandleLambdaEvent, MoesifOptions()))
}
```

### 2. Enter Moesif Application Id
Your Moesif Application Id can be found in the [_Moesif Portal_](https://www.moesif.com/).
After signing up for a Moesif account, your Moesif Application Id will be displayed during the onboarding steps. 

You can always find your Moesif Application Id at any time by logging 
into the [_Moesif Portal_](https://www.moesif.com/), click on the top right menu,
 and then clicking _Installation_.

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