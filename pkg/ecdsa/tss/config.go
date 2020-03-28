package tss

// Config contains configuration for tss protocol execution.
type Config struct {
	// Concurrency level for pre-parameters generation in tss-lib.
	PreParamsGenerationConcurrency int
}
