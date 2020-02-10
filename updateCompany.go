package moesifawslambda

import (
	"log"
	models "github.com/moesif/moesifapi-go/models"
)

 // Update Company
 func UpdateCompanyAsync(company *models.CompanyModel, configurationOption map[string]interface{}) {
	 
	// Call the function to initialize the moesif client and moesif options
	if apiClient == nil {
		moesifClient(configurationOption)
	}

	// Update company profile
	errUpdateCompany := apiClient.UpdateCompany(company)
	// Log the message
	if errUpdateCompany != nil {
		log.Fatalf("Error while updating company: %s.\n", errUpdateCompany.Error())
	} else {
		log.Println("Company updated successfully")
	}
 }

 // Update Companies Batch
 func UpdateCompaniesBatchAsync(companies []*models.CompanyModel, configurationOption map[string]interface{}) {
	 
	// Call the function to initialize the moesif client and moesif options
	if apiClient == nil {
		moesifClient(configurationOption)
	}

	// Update company profiles
	errUpdateCompaniesBatch := apiClient.UpdateCompaniesBatch(companies)
	// Log the message
	if errUpdateCompaniesBatch != nil {
		log.Fatalf("Error while updating companies in batch: %s.\n", errUpdateCompaniesBatch.Error())
	} else {
		log.Println("Companies updated successfully")
	}
 }
