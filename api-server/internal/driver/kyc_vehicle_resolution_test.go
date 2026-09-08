package driver

import (
	"testing"

	"github.com/workspace/ride-platform/pkg/documents"
)

func TestUploadDocument_RequiresVehicleTypes(t *testing.T) {
	// Person-level documents: shared across every vehicle, carried with vehicle_id = NULL
	personDocs := []string{
		documents.NationalIDFront,
		documents.NationalIDBack,
		documents.LicenceFront,
		documents.LicenceBack,
		documents.Selfie,
		"PROFILE_SELFIE", // alias for Selfie
	}
	for _, doc := range personDocs {
		if documents.RequiresVehicle(doc) {
			t.Errorf("expected person-level document %s to NOT require vehicle_id", doc)
		}
	}

	// Vehicle-level documents: describe a specific vehicle, MUST have vehicle_id
	vehicleDocs := []string{
		documents.VehicleInsurance,
		documents.VehicleInsuranceBack,
		documents.VehicleAuthorization,
		documents.VehicleAuthorizationBack,
	}
	for _, doc := range vehicleDocs {
		if !documents.RequiresVehicle(doc) {
			t.Errorf("expected vehicle-level document %s to require vehicle_id", doc)
		}
	}
}
