package models

import (
	"testing"

	"test_utils"

	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
	// "google.golang.org/appengine/log"
	"gopkg.in/go-playground/validator.v9"
)

func TestOrganizationCRUD(t *testing.T) {
	opt := &aetest.Options{StronglyConsistentDatastore: true}
	inst, err := aetest.NewInstance(opt)
	assert.NoError(t, err)
	defer inst.Close()

	req, err := inst.NewRequest("GET", "/", nil)
	if !assert.NoError(t, err) {
		inst.Close()
		return
	}
	ctx := appengine.NewContext(req)

	test_utils.ClearDatastore(t, ctx, "Organizations")

	detectErrorFor := func(errors validator.ValidationErrors, field string) validator.FieldError {
		for _, err := range errors {
			if err.StructField() == field {
				return err
			}
		}
		return nil
	}

	org := &Organization{}
	err = org.Validate()

	assert.Error(t, err)
	errors := err.(validator.ValidationErrors)

	fields := []string{
		"Name",
	}
	for _, field := range fields {
		err := detectErrorFor(errors, field)
		if assert.NotNil(t, err) {
			assert.Equal(t, "required", err.ActualTag())
		}
	}

	org1 := &Organization{
		Name: "org01",
	}
	err = org1.Create(ctx)
	assert.NoError(t, err)
	key, err := datastore.DecodeKey(org1.ID)
	assert.NoError(t, err)

	org2 := &Organization{}
	err = datastore.Get(ctx, key, org2)
	assert.NoError(t, err)

	// Find
	org3, err := GlobalOrganizationAccessor.Find(ctx, org1.ID)
	assert.NoError(t, err)
	assert.Equal(t, org1.Name, org3.Name)

	// Validate before update
	org1.Name = ""
	err = org1.Update(ctx)
	assert.Error(t, err)

	// Update
	org1.Name = "Org01"
	err = org1.Update(ctx)
	assert.NoError(t, err)

	// GetAll
	orgs, err := GlobalOrganizationAccessor.All(ctx)
	assert.NoError(t, err)
	if len(orgs) != 1 {
		t.Fatalf("len(orgs) expects %v but was %v\n", 1, len(orgs))
	}
	assert.Equal(t, org1.Name, orgs[0].Name)

	err = org1.Destroy(ctx)
	assert.NoError(t, err)
}
