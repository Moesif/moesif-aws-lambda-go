package moesifawslambda

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	moesifapi "github.com/moesif/moesifapi-go"
	models "github.com/moesif/moesifapi-go/models"
)

// Global variable
var (
	apiClient              moesifapi.API
	debug                  bool
	logBody                bool
	disableCaptureOutgoing bool
	logBodyOutgoing        bool
	moesifOption           map[string]interface{}
)

// Start Capture Outgoing Request
func StartCaptureOutgoing(configurationOption map[string]interface{}) {
	// Call the function to initialize the moesif client and moesif options
	if apiClient == nil {
		// Set the Capture_Outoing_Requests to true to capture outgoing request
		configurationOption["Capture_Outoing_Requests"] = true
		moesifOption = configurationOption
		moesifClient(moesifOption)
	}

	if debug {
		log.Println("Start Capturing outgoing requests")
	}
	// Enable logBody by default
	logBodyOutgoing = true
	// Try to fetch the disableTransactionId from the option
	if isEnabled, found := moesifOption["Log_Body_Outgoing"].(bool); found {
		logBodyOutgoing = isEnabled
	}

	http.DefaultTransport = DefaultTransport
}

// Function to update User
func UpdateUser(user *models.UserModel, configurationOption map[string]interface{}) {
	UpdateUserAsync(user, configurationOption)
}

// Function to update Users in batch
func UpdateUsersBatch(users []*models.UserModel, configurationOption map[string]interface{}) {
	UpdateUsersBatchAsync(users, configurationOption)
}

// Function to update User
func UpdateCompany(company *models.CompanyModel, configurationOption map[string]interface{}) {
	UpdateCompanyAsync(company, configurationOption)
}

// Function to update Users in batch
func UpdateCompaniesBatch(companies []*models.CompanyModel, configurationOption map[string]interface{}) {
	UpdateCompaniesBatchAsync(companies, configurationOption)
}

// Initialize the client
func moesifClient(moesifOption map[string]interface{}) {
	// var apiEndpoint string
	// var batchSize int
	// var eventQueueSize int
	// var timerWakeupSeconds int

	applicationId := os.Getenv("MOESIF_APPLICATION_ID")
	api := moesifapi.NewAPI(applicationId) // , &apiEndpoint, eventQueueSize, batchSize, timerWakeupSeconds
	apiClient = api

	//  Disable debug by default
	debug = false
	// Try to fetch the debug from the option
	if isDebug, found := moesifOption["Debug"].(bool); found {
		debug = isDebug
	}

	// Enable logBody by default
	logBody = true

	// Try to fetch the logBody from the option
	if isEnabled, found := moesifOption["Log_Body"].(bool); found {
		logBody = isEnabled
	}
}

func getUserId(request events.APIGatewayProxyRequest, response events.APIGatewayProxyResponse) *string {
	var username string
	if _, found := moesifOption["Identify_User"]; found {
		username = moesifOption["Identify_User"].(func(events.APIGatewayProxyRequest, events.APIGatewayProxyResponse) string)(request, response)
		return &username
	} else {
		if len(request.RequestContext.Identity.CognitoIdentityID) > 0 {
			return &request.RequestContext.Identity.CognitoIdentityID
		} else {
			return nil
		}
	}
}

func getUserIdV2HTTP(request events.APIGatewayV2HTTPRequest, response events.APIGatewayV2HTTPResponse) *string {
	var username string
	if _, found := moesifOption["Identify_User"]; found {
		username = moesifOption["Identify_User"].(func(events.APIGatewayV2HTTPRequest, events.APIGatewayV2HTTPResponse) string)(request, response)
		return &username
	} else {
		switch (request.RequestContext.Authorizer != nil) && (request.RequestContext.Authorizer.IAM != nil) {
		case true:
			identity := request.RequestContext.Authorizer.IAM
			if len(identity.CognitoIdentity.IdentityID) > 0 {
				return &request.RequestContext.Authorizer.IAM.CognitoIdentity.IdentityID
			}
		case false:
			return nil
		}
		return nil
	}
}

func sendMoesifAsyncV2HTTP(request events.APIGatewayV2HTTPRequest, response events.APIGatewayV2HTTPResponse, configurationOption map[string]interface{}) {

	// Api Version
	var apiVersion *string = nil
	if isApiVersion, found := moesifOption["Api_Version"].(string); found {
		apiVersion = &isApiVersion
	}

	// Get Metadata
	var metadata map[string]interface{} = nil
	if _, found := moesifOption["Get_Metadata"]; found {
		metadata = moesifOption["Get_Metadata"].(func(events.APIGatewayV2HTTPRequest, events.APIGatewayV2HTTPResponse) map[string]interface{})(request, response)
	}

	// Get User
	var userId *string
	userId = getUserIdV2HTTP(request, response)

	// Get Company
	var companyId string
	if _, found := moesifOption["Identify_Company"]; found {
		companyId = moesifOption["Identify_Company"].(func(events.APIGatewayV2HTTPRequest, events.APIGatewayV2HTTPResponse) string)(request, response)
	}

	// Get Session Token
	var sessionToken string
	if _, found := moesifOption["Get_Session_Token"]; found {
		sessionToken = moesifOption["Get_Session_Token"].(func(events.APIGatewayV2HTTPRequest, events.APIGatewayV2HTTPResponse) string)(request, response)
	}

	// Prepare Moesif Event
	moesifEvent := prepareEventV2HTTP(request, response, apiVersion, userId, companyId, sessionToken, metadata)
	jsonEvent, _ := json.Marshal(moesifEvent)
	fmt.Println("HERE'S THE MOESIF EVENT MODEL:>>>>")
	fmt.Println(string(jsonEvent))

	// Should skip
	shouldSkip := false
	if _, found := moesifOption["Should_Skip"]; found {
		shouldSkip = moesifOption["Should_Skip"].(func(events.APIGatewayV2HTTPRequest, events.APIGatewayV2HTTPResponse) bool)(request, response)
	}

	if shouldSkip {
		if debug {
			log.Printf("Skip sending the event to Moesif")
		}
	} else {
		if debug {
			log.Printf("Sending the event to Moesif")
		}

		if _, found := moesifOption["Mask_Event_Model"]; found {
			moesifEvent = moesifOption["Mask_Event_Model"].(func(models.EventModel) models.EventModel)(moesifEvent)
		}

		// Call the function to send event to Moesif
		_, err := apiClient.CreateEvent(&moesifEvent)

		if err != nil {
			log.Fatalf("Error while sending event to Moesif: %s.\n", err.Error())
		}

		if debug {
			log.Printf("Successfully sent event to Moesif")
		}
	}
}

func sendMoesifAsync(request events.APIGatewayProxyRequest, response events.APIGatewayProxyResponse, configurationOption map[string]interface{}) {

	// Api Version
	var apiVersion *string = nil
	if isApiVersion, found := moesifOption["Api_Version"].(string); found {
		apiVersion = &isApiVersion
	}

	// Get Metadata
	var metadata map[string]interface{} = nil
	if _, found := moesifOption["Get_Metadata"]; found {
		metadata = moesifOption["Get_Metadata"].(func(events.APIGatewayProxyRequest, events.APIGatewayProxyResponse) map[string]interface{})(request, response)
	}

	// Get User
	var userId *string
	userId = getUserId(request, response)

	// Get Company
	var companyId string
	if _, found := moesifOption["Identify_Company"]; found {
		companyId = moesifOption["Identify_Company"].(func(events.APIGatewayProxyRequest, events.APIGatewayProxyResponse) string)(request, response)
	}

	// Get Session Token
	var sessionToken string
	if _, found := moesifOption["Get_Session_Token"]; found {
		sessionToken = moesifOption["Get_Session_Token"].(func(events.APIGatewayProxyRequest, events.APIGatewayProxyResponse) string)(request, response)
	}

	// Prepare Moesif Event
	moesifEvent := prepareEvent(request, response, apiVersion, userId, companyId, sessionToken, metadata)

	// Should skip
	shouldSkip := false
	if _, found := moesifOption["Should_Skip"]; found {
		shouldSkip = moesifOption["Should_Skip"].(func(events.APIGatewayProxyRequest, events.APIGatewayProxyResponse) bool)(request, response)
	}

	if shouldSkip {
		if debug {
			log.Printf("Skip sending the event to Moesif")
		}
	} else {
		if debug {
			log.Printf("Sending the event to Moesif")
		}

		if _, found := moesifOption["Mask_Event_Model"]; found {
			moesifEvent = moesifOption["Mask_Event_Model"].(func(models.EventModel) models.EventModel)(moesifEvent)
		}

		// Call the function to send event to Moesif
		_, err := apiClient.CreateEvent(&moesifEvent)

		if err != nil {
			log.Fatalf("Error while sending event to Moesif: %s.\n", err.Error())
		}

		if debug {
			log.Printf("Successfully sent event to Moesif")
		}
	}
}

func MoesifLogger(f func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error), configurationOption map[string]interface{}) func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		response, err := f(ctx, request)
		// Call the function to initialize the moesif client and moesif options
		if apiClient == nil {
			moesifOption = configurationOption
			moesifClient(moesifOption)
		}
		sendMoesifAsync(request, response, configurationOption)
		return response, err
	}
}

func MoesifLoggerV2HTTP(f func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error), configurationOption map[string]interface{}) func(context.Context, events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		response, err := f(ctx, request)
		// Call the function to initialize the moesif client and moesif options
		if apiClient == nil {
			moesifOption = configurationOption
			moesifClient(moesifOption)
		}
		sendMoesifAsyncV2HTTP(request, response, configurationOption)
		return response, err
	}
}
