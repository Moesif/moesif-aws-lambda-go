package moesifawslambda

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func MoesifOptions() map[string]interface{} {
	var moesifOptions = map[string]interface{}{
		"Application_Id":    "",
		"Api_Version":       "1.0.0",
		"Debug":             true,
		"Log_Body":          true,
		"Log_Body_Outgoing": true,
	}
	return moesifOptions
}

func HandleLambdaEvent(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       request.Body,
		StatusCode: 200,
		Headers: map[string]string{
			"RspHeader1":     "RspHeaderValue1",
			"Content-Type":   "application/json",
			"Content-Length": "1000",
		},
	}, nil
}

func HandleLambdaEventV2HTTP(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"RspHeader1":     "TheYearofDesktopLinux",
				"Content-Type":   "application/json",
				"Content-Length": "1000",
			},
			MultiValueHeaders: map[string][]string{
				"X-Forwarded-For":   {"127.0.0.1, 127.0.0.2"},
				"X-Forwarded-Port":  {"443"},
				"X-Forwarded-Proto": {"https"},
			},
			Body:            request.Body,
			IsBase64Encoded: false,
			Cookies:         []string{"cookie1", "cookie2"},
		},
		nil
}

// Generates mock `events.APIGatewayProxyRequest` request objects.
func generateProxyReq(body []byte, isb64Encoded bool) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		Resource:                        "foo/bar",
		Path:                            "foo/bar/dev",
		HTTPMethod:                      "POST",
		Headers:                         map[string]string{"Content-Type": "application/json"},
		MultiValueHeaders:               map[string][]string{"X-Forwarded-Proto": {"https"}},
		QueryStringParameters:           map[string]string{"foo": "bar"},
		MultiValueQueryStringParameters: map[string][]string{"foo": {"bar"}},
		PathParameters:                  map[string]string{"proxy": "/path/to/resource"},
		StageVariables:                  map[string]string{"baz": "bar"},
		RequestContext:                  events.APIGatewayProxyRequestContext{},
		Body:                            string(body),
		IsBase64Encoded:                 isb64Encoded,
	}
}

// Generates mock `events.APIGatewayV2HTTPRequest` request objects.
func generateProxyReqV2HTTP(body []byte, isb64Encoded bool) events.APIGatewayV2HTTPRequest {
	return events.APIGatewayV2HTTPRequest{
		Version:               "2.0",
		RouteKey:              "$default",
		RawPath:               "/path/to/foo",
		RawQueryString:        "parameter1=value1&parameter1=value2&parameter2=value",
		Cookies:               []string{"cookie1", "cookie2"},
		Headers:               map[string]string{"Header1": "value1", "Header2": "value1,value2"},
		QueryStringParameters: map[string]string{"foo": "bar"},
		PathParameters:        map[string]string{"proxy": "/path/to/resource"},
		RequestContext:        events.APIGatewayV2HTTPRequestContext{},
		StageVariables:        map[string]string{"baz": "bar"},
		Body:                  string(body),
		IsBase64Encoded:       isb64Encoded,
	}
}

// This function mocks a portion of `prepareEvent` that processes the request body
// and calls `processBody` accordingly.
// Returns the same as `processBody`.
func mockPrepareEvent(request events.APIGatewayProxyRequest) (interface{}, string) {
	var transformReqBody interface{} = nil
	var transferEncoding string = "json"

	if logBody && len(request.Body) != 0 {
		if request.IsBase64Encoded && isBase64String(request.Body) {
			transformReqBody = request.Body
			transferEncoding = "base64"
		} else {
			transformReqBody, transferEncoding = processBody(request.Body)
		}
	}
	return transformReqBody, transferEncoding
}

// This function mocks a portion of `prepareEventV2HTTP` that processes the request body
// and calls `processBody` accordingly.
// Returns the same as `processBody`.
func mockPrepareEventV2HTTP(request events.APIGatewayV2HTTPRequest) (interface{}, string) {
	var transformReqBody interface{} = nil
	var transferEncoding string = "json"

	if logBody && len(request.Body) != 0 {
		if request.IsBase64Encoded {
			switch isBase64String(request.Body) {
			case true:
				transformReqBody = request.Body
				transferEncoding = "base64"
			case false:
				// Meaning body isn't a valid base64-encoded string despite
				// `IsBase64Encoded``  being `true`.
				// So we try to pass it on to `processBody`. If the body is not a
				// valid JSON, we encode it to base64.
				transformReqBody, transferEncoding = processBody(request.Body)
				// We want to set `transferEncoding` to empty string if `transferEncoding`
				// is JSON. This parallels our implementation in Node.js Lambda middleware.
				if transferEncoding == "json" {
					transferEncoding = ""
				}
			}
		} else {
			transformReqBody, transferEncoding = processBody(request.Body)
		}
	}
	return transformReqBody, transferEncoding
}

func TestProcessBody(t *testing.T) {

	var proxyReqWithJsonStrBody = generateProxyReq([]byte(`{"foo": "bar"}`), false)
	var proxyReqWithBase64StrBody = generateProxyReq([]byte(`eyJmb28iOiAiYmFyIn0=`), true)
	var proxyReqWithInvalidBase64StrBody = generateProxyReq([]byte(`{"foo": "bar"}`), true)

	type expected struct {
		expectedBody             interface{}
		expectedTransferEncoding string
	}

	var testcases = []struct {
		in  events.APIGatewayProxyRequest
		out expected
	}{
		{proxyReqWithJsonStrBody, expected{expectedBody: map[string]interface{}{"foo": "bar"}, expectedTransferEncoding: "json"}},
		{proxyReqWithBase64StrBody, expected{expectedBody: "eyJmb28iOiAiYmFyIn0=", expectedTransferEncoding: "base64"}},
		{proxyReqWithInvalidBase64StrBody, expected{expectedBody: map[string]interface{}{"foo": "bar"}, expectedTransferEncoding: "json"}},
	}

	for _, tt := range testcases {

		res := MoesifLogger(HandleLambdaEvent, MoesifOptions())

		result, err := res(context.Background(), tt.in)
		if err != nil {
			t.Logf("encountered error\n")
			t.Logf("===========\n")
			t.Log(result)
			t.Logf("\n===========\n")
		}

		parsedBody, transferEncoding := mockPrepareEvent(tt.in)

		if transferEncoding != tt.out.expectedTransferEncoding {
			t.Errorf("got %v, want %v", transferEncoding, tt.out.expectedTransferEncoding)
		}
		if !reflect.DeepEqual(parsedBody, tt.out.expectedBody) {
			t.Errorf("got %v, want %v", parsedBody, tt.out.expectedBody)
		}

	}

}

func TestProcessBodyV2HTTP(t *testing.T) {

	var proxyReqWithJsonStrBody = generateProxyReqV2HTTP([]byte(`{"foo": "bar"}`), false)
	var proxyReqWithBase64StrBody = generateProxyReqV2HTTP([]byte(`eyJmb28iOiAiYmFyIn0=`), true)
	var proxyReqWithInvalidBase64StrBody = generateProxyReqV2HTTP([]byte(`{"foo": "bar"}`), true)

	type expected struct {
		expectedBody             interface{}
		expectedTransferEncoding string
	}

	var testcases = []struct {
		in  events.APIGatewayV2HTTPRequest
		out expected
	}{
		{proxyReqWithJsonStrBody, expected{expectedBody: map[string]interface{}{"foo": "bar"}, expectedTransferEncoding: "json"}},
		{proxyReqWithBase64StrBody, expected{expectedBody: "eyJmb28iOiAiYmFyIn0=", expectedTransferEncoding: "base64"}},
		{proxyReqWithInvalidBase64StrBody, expected{expectedBody: map[string]interface{}{"foo": "bar"}, expectedTransferEncoding: ""}},
	}

	for _, tt := range testcases {

		res := MoesifLoggerV2HTTP(HandleLambdaEventV2HTTP, MoesifOptions())

		result, err := res(context.Background(), tt.in)
		if err != nil {
			t.Logf("encountered error\n")
			t.Logf("===========\n")
			t.Log(result)
			t.Logf("\n===========\n")
		}

		parsedBody, transferEncoding := mockPrepareEventV2HTTP(tt.in)

		if transferEncoding != tt.out.expectedTransferEncoding {
			t.Errorf("got %v, want %v", transferEncoding, tt.out.expectedTransferEncoding)
		}
		if !reflect.DeepEqual(parsedBody, tt.out.expectedBody) {
			t.Errorf("got %v, want %v", parsedBody, tt.out.expectedBody)
		}

	}

}

func TestPrepareRequestURIV2HTTP(t *testing.T) {
	var uri = prepareRequestURIV2HTTP(generateProxyReqV2HTTP([]byte(`{"name": "Alex"}`), false))

	var expected = "http://localhost/path/to/foo?parameter1=value1&parameter1=value2&parameter2=value"

	if uri != expected {
		t.Errorf("got %v, want %v", uri, expected)
	}
}
