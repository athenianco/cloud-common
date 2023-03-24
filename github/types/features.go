package types

const (
	// FeatureGHE is set for accounts based on Github Enterprise instance, instead of github.com.
	FeatureGHE = Feature("athenian.github.ghe")
	// FeatureNoConsistency disables consistency checks for an account.
	FeatureNoConsistency         = Feature("athenian.github.no_consistency")
	FeatureNoCICheckRun          = Feature("athenian.github.drop_node_type.CheckRun")
	FeatureNoCICheckSuite        = Feature("athenian.github.drop_node_type.CheckSuite")
	FeatureNoCIStatus            = Feature("athenian.github.drop_node_type.Status")
	FeatureNoCIStatusCheckRollup = Feature("athenian.github.drop_node_type.StatusCheckRollup")
	FeatureNoCIStatusContext     = Feature("athenian.github.drop_node_type.StatusContext")
)
