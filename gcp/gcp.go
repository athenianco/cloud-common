package gcp

import (
	"github.com/athenianco/cloud-common/envs"
)

// projectID is set from the GCP_PROJECT (which is automatically set by the Cloud Functions runtime)
// or ATHENIAN_GCP_PROJECT environment variable.
var projectID = envs.OneOfEnvs("GCP_PROJECT", "ATHENIAN_GCP_PROJECT")

// ProjectID returns a GCP project ID, if it's set.
func ProjectID() string {
	return projectID
}
