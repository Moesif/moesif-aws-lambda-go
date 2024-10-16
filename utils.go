package moesifawslambda

import (
	b64 "encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/aws/aws-lambda-go/events"
	models "github.com/moesif/moesifapi-go/models"
)

const Base64 string = "^[A-Za-z0-9+/]+={0,2}$"

func prepareRequestURI(request events.APIGatewayProxyRequest) string {
	var uri string
	if forwardedProtoHeader, found := request.Headers["X-Forwarded-Proto"]; found {
		uri = forwardedProtoHeader
	} else {
		uri = "http"
	}

	uri += "://"

	if hostHeader, found := request.Headers["Host"]; found {
		uri += hostHeader
	} else {
		uri += "localhost"
	}

	if pathHeader, found := request.Headers["path"]; found {
		uri += pathHeader
	} else {
		uri += "/"
	}

	if len(request.MultiValueQueryStringParameters) > 0 {
		queryString := ""
		for q, l := range request.MultiValueQueryStringParameters {
			for _, v := range l {
				if queryString != "" {
					queryString += "&"
				}
				queryString += url.QueryEscape(q) + "=" + url.QueryEscape(v)
			}
		}
		uri += "?" + queryString
	} else if len(request.QueryStringParameters) > 0 {
		queryString := ""
		for q := range request.QueryStringParameters {
			if queryString != "" {
				queryString += "&"
			}
			queryString += url.QueryEscape(q) + "=" + url.QueryEscape(request.QueryStringParameters[q])
		}
		uri += "?" + queryString
	}
	return uri
}

func prepareRequestURIV2HTTP(request events.APIGatewayV2HTTPRequest) string {
	var uri string
	if forwardedProtoHeader, found := request.Headers["x-forwarded-proto"]; found {
		uri = forwardedProtoHeader
	} else {
		uri = "http"
	}

	uri += "://"

	if hostHeader, found := request.Headers["host"]; found {
		uri += hostHeader
	} else {
		uri += "localhost"
	}

	if len(request.RawPath) > 0 {
		uri += request.RawPath
	} else {
		uri = "/"
	}

	if len(request.RawQueryString) > 0 {
		uri += "?" + request.RawQueryString
	}
	return uri
}

func isBase64String(str string) bool {
	b64Regex, err := regexp.Compile(Base64)
	if err != nil {
		return false
	}
	return b64Regex.MatchString(str)
}

func processBody(body string) (interface{}, string) {
	var parsedBody interface{}
	var transferEncoding string

	parsedBody = nil
	transferEncoding = "json"
	if jsonMarshalErr := json.Unmarshal([]byte(body), &parsedBody); jsonMarshalErr != nil {
		if debug {
			log.Printf("About to parse request body as base64 ")
		}
		parsedBody = b64.StdEncoding.EncodeToString([]byte(body))
		transferEncoding = "base64"
		if debug {
			log.Printf("Parsed request body as base64 - %s", parsedBody)
		}
	}
	return parsedBody, transferEncoding
}

func processHeaders(headers map[string]string) map[string]string {
	// Check if the headers are empty
	if len(headers) == 0 {
		var emptyHeaders = map[string]string{}
		return emptyHeaders
	} else {
		return headers
	}
}

func defaultSourceIp(request events.APIGatewayProxyRequest) *string {
	if len(request.RequestContext.Identity.SourceIP) > 0 {
		return &request.RequestContext.Identity.SourceIP
	} else {
		return nil
	}
}

func defaultSourceIpV2HTTP(request events.APIGatewayV2HTTPRequest) *string {
	if len(request.RequestContext.HTTP.SourceIP) > 0 {
		return &request.RequestContext.HTTP.SourceIP
	} else {
		return nil
	}
}

func prepareEvent(request events.APIGatewayProxyRequest, response events.APIGatewayProxyResponse, apiVersion *string, userId *string, companyId string, sessionToken string, metadata map[string]interface{}) models.EventModel {

	reqTime := time.Now().UTC()
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

	var transformReqHeaders = make(map[string][]string)
	for key, value := range request.Headers {
		transformReqHeaders[key] = []string{value}
	}

	eventRequestModel := models.EventRequestModel{
		Time:             &reqTime,
		Uri:              prepareRequestURI(request),
		Verb:             request.HTTPMethod,
		ApiVersion:       apiVersion,
		IpAddress:        getClientIp(transformReqHeaders, defaultSourceIp(request)),
		Headers:          processHeaders(request.Headers),
		Body:             &transformReqBody,
		TransferEncoding: &transferEncoding,
	}

	rspTime := time.Now().UTC()

	var transformRespBody interface{}
	transferEncoding = "json"

	if logBody && len(response.Body) != 0 {
		if response.IsBase64Encoded {
			transformRespBody = response.Body
			transferEncoding = "base64"
		} else {
			transformRespBody, transferEncoding = processBody(response.Body)
		}
	}

	eventResponseModel := models.EventResponseModel{
		Time:             &rspTime,
		Status:           response.StatusCode,
		IpAddress:        nil,
		Headers:          processHeaders(response.Headers),
		Body:             &transformRespBody,
		TransferEncoding: &transferEncoding,
	}

	direction := "Incoming"
	weight := 1

	event := models.EventModel{
		Request:      eventRequestModel,
		Response:     eventResponseModel,
		SessionToken: &sessionToken,
		Tags:         nil,
		UserId:       userId,
		CompanyId:    &companyId,
		Metadata:     &metadata,
		Direction:    &direction,
		Weight:       &weight,
	}
	return event
}

func prepareEventV2HTTP(request events.APIGatewayV2HTTPRequest, response events.APIGatewayV2HTTPResponse, apiVersion *string, userId *string, companyId string, sessionToken string, metadata map[string]interface{}) models.EventModel {

	reqTime := time.Now().UTC()
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

	var transformReqHeaders = make(map[string][]string)
	for key, value := range request.Headers {
		transformReqHeaders[key] = []string{value}
	}

	eventRequestModel := models.EventRequestModel{
		Time:             &reqTime,
		Uri:              prepareRequestURIV2HTTP(request),
		Verb:             request.RequestContext.HTTP.Method,
		ApiVersion:       apiVersion,
		IpAddress:        getClientIp(transformReqHeaders, defaultSourceIpV2HTTP(request)),
		Headers:          processHeaders(request.Headers),
		Body:             &transformReqBody,
		TransferEncoding: &transferEncoding,
	}

	rspTime := time.Now().UTC()

	var transformRespBody interface{}
	transferEncoding = "json"

	if logBody && len(response.Body) != 0 {
		if response.IsBase64Encoded {
			transformRespBody = response.Body
			transferEncoding = "base64"
		} else {
			transformRespBody, transferEncoding = processBody(response.Body)
		}
	}

	eventResponseModel := models.EventResponseModel{
		Time:             &rspTime,
		Status:           response.StatusCode,
		IpAddress:        nil,
		Headers:          processHeaders(response.Headers),
		Body:             &transformRespBody,
		TransferEncoding: &transferEncoding,
	}

	direction := "Incoming"
	weight := 1

	event := models.EventModel{
		Request:      eventRequestModel,
		Response:     eventResponseModel,
		SessionToken: &sessionToken,
		Tags:         nil,
		UserId:       userId,
		CompanyId:    &companyId,
		Metadata:     &metadata,
		Direction:    &direction,
		Weight:       &weight,
	}
	return event
}

// Send Outgoing Event to Moesif
func sendMoesifOutgoingAsync(request *http.Request, reqTime time.Time, apiVersion *string, reqBody interface{}, reqEncoding *string,
	rspTime time.Time, respStatus int, respHeader http.Header, respBody interface{}, respEncoding *string, userId *string,
	companyId *string, sessionToken *string, metadata map[string]interface{}, direction *string, weight *int) {

	// Get Client Ip
	ip := getClientIp(request.Header, nil)

	// Prepare request model
	event_request := models.EventRequestModel{
		Time:             &reqTime,
		Uri:              request.URL.Scheme + "://" + request.Host + request.URL.Path,
		Verb:             request.Method,
		ApiVersion:       apiVersion,
		IpAddress:        ip,
		Headers:          request.Header,
		Body:             &reqBody,
		TransferEncoding: reqEncoding,
	}

	// Prepare response model
	event_response := models.EventResponseModel{
		Time:             &rspTime,
		Status:           respStatus,
		IpAddress:        nil,
		Headers:          respHeader,
		Body:             respBody,
		TransferEncoding: respEncoding,
	}

	// Prepare the event model
	event := models.EventModel{
		Request:      event_request,
		Response:     event_response,
		SessionToken: sessionToken,
		Tags:         nil,
		UserId:       userId,
		CompanyId:    companyId,
		Metadata:     metadata,
		Direction:    direction,
		Weight:       weight,
	}

	// Send event to moesif
	_, err := apiClient.CreateEvent(&event)

	// Log the message
	if err != nil {
		log.Fatalf("Error while sending event to Moesif: %s.\n", err.Error())
	}

	if debug {
		log.Printf("Successfully sent outgoing event to Moesif")
	}
}
