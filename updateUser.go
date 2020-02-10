
package moesifawslambda

import (
	"log"
	models "github.com/moesif/moesifapi-go/models"
)

// Update User
func UpdateUserAsync(user *models.UserModel, configurationOption map[string]interface{}) {
	 
	// Call the function to initialize the moesif client and moesif options
	if apiClient == nil {
		moesifClient(configurationOption)
	}

	// Update user profile
	errUpdateUser := apiClient.UpdateUser(user)
	// Log the message
	if errUpdateUser != nil {
		log.Fatalf("Error while updating user: %s.\n", errUpdateUser.Error())
	} else {
		log.Println("User updated successfully")
	}
 }

 // Update Users Batch
 func UpdateUsersBatchAsync(users []*models.UserModel, configurationOption map[string]interface{}) {
	 
	// Call the function to initialize the moesif client and moesif options
	if apiClient == nil {
		moesifClient(configurationOption)
	}

	// Update user profiles
	errUpdateUserBatch := apiClient.UpdateUsersBatch(users)
	// Log the message
	if errUpdateUserBatch != nil {
		log.Fatalf("Error while updating users in batch: %s.\n", errUpdateUserBatch.Error())
	} else {
		log.Println("Users updated successfully")
	}
 }