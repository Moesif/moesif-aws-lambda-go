package moesifawslambda

import (
	"net/url"
	"log"
	"time"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	models "github.com/moesif/moesifapi-go/models"
	b64 "encoding/base64"
)

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

func processBody(body string) (interface{}, string) {
	var parsedBody interface{}
	var transferEncoding string

	parsedBody = nil
	transferEncoding = "json"
	if logBody {
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
	}
	return parsedBody, transferEncoding
}

func processHeaders(headers map[string] string) map[string] string {
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

func prepareEvent(request events.APIGatewayProxyRequest, response events.APIGatewayProxyResponse, apiVersion *string, userId string, companyId string, sessionToken string, metadata map[string]interface{}) models.EventModel {

	reqTime := time.Now().UTC()
	transformReqBody, transferEncoding := processBody(request.Body)

	eventRequestModel := models.EventRequestModel{
		Time:       &reqTime,
		Uri:        prepareRequestURI(request),
		Verb:       request.HTTPMethod,
		ApiVersion: apiVersion,
		IpAddress: getClientIp(request.Headers, defaultSourceIp(request)),
		Headers: processHeaders(request.Headers),
		Body: &transformReqBody,
		TransferEncoding: &transferEncoding,
	}

	rspTime := time.Now().UTC()
	transformRespBody, transferEncoding := processBody(response.Body)

	eventResponseModel := models.EventResponseModel{
		Time:      &rspTime,
		Status:    response.StatusCode,
		IpAddress: nil,
		Headers: processHeaders(response.Headers),
		Body: &transformRespBody,
		TransferEncoding: &transferEncoding,
	}

	event := models.EventModel{
		Request:      eventRequestModel,
		Response:     eventResponseModel,
		SessionToken: &sessionToken,
		Tags:         nil,
		UserId:       &userId,
		CompanyId:    &companyId,
		Metadata: 	  &metadata,
	}
	return event
}
