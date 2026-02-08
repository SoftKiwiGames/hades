package loader

import (
	"fmt"
	"strings"

	"github.com/SoftKiwiGames/hades/hades/schema"
)

// ValidateEnvContract validates that provided environment variables satisfy the job's contract
func ValidateEnvContract(job *schema.Job, provided map[string]string) error {
	// Check that all required env vars are provided
	for name, envDef := range job.Env {
		// Check if user tried to override HADES_* variables
		if strings.HasPrefix(name, "HADES_") {
			return fmt.Errorf("job cannot define HADES_* environment variables: %s", name)
		}

		// If no default, the env var is required
		if envDef.Default == "" {
			if _, ok := provided[name]; !ok {
				return fmt.Errorf("required environment variable %q not provided", name)
			}
		}
	}

	// Check that no unknown env vars are provided
	for name := range provided {
		// HADES_* vars will be added by runtime, so they're expected
		if strings.HasPrefix(name, "HADES_") {
			return fmt.Errorf("user cannot provide HADES_* environment variables: %s", name)
		}

		// Check if this env var is defined in the job
		if _, ok := job.Env[name]; !ok {
			return fmt.Errorf("unknown environment variable %q (not defined in job env contract)", name)
		}
	}

	return nil
}

// MergeEnv merges environment variables with priority: provided > defaults
// Returns the complete environment with all required and optional variables
func MergeEnv(job *schema.Job, provided map[string]string) map[string]string {
	result := make(map[string]string)

	// First, add all job defaults
	for name, envDef := range job.Env {
		if envDef.Default != "" {
			result[name] = envDef.Default
		}
	}

	// Then, override with provided values
	for name, value := range provided {
		result[name] = value
	}

	return result
}

// ValidateStepEnv validates step-level environment variables against the job's contract
func ValidateStepEnv(file *schema.File, planName string, stepIdx int) error {
	plan := file.Plans[planName]
	if stepIdx >= len(plan.Steps) {
		return fmt.Errorf("step index %d out of range", stepIdx)
	}

	step := plan.Steps[stepIdx]
	job, ok := file.Jobs[step.Job]
	if !ok {
		return fmt.Errorf("job %q not found", step.Job)
	}

	// Validate step-level env vars
	for envName := range step.Env {
		if strings.HasPrefix(envName, "HADES_") {
			return fmt.Errorf("step %q: cannot define HADES_* environment variables: %s", step.Name, envName)
		}

		if _, ok := job.Env[envName]; !ok {
			return fmt.Errorf("step %q: unknown environment variable %q (not defined in job %q)", step.Name, envName, step.Job)
		}
	}

	return nil
}

// ValidatePlanEnv validates all environment variables in a plan
func ValidatePlanEnv(file *schema.File, planName string, cliEnv map[string]string) error {
	plan, ok := file.Plans[planName]
	if !ok {
		return fmt.Errorf("plan %q not found", planName)
	}

	// Check CLI env vars don't start with HADES_
	for name := range cliEnv {
		if strings.HasPrefix(name, "HADES_") {
			return fmt.Errorf("cannot provide HADES_* environment variables via CLI: %s", name)
		}
	}

	// Validate each step
	for i, step := range plan.Steps {
		job, ok := file.Jobs[step.Job]
		if !ok {
			return fmt.Errorf("step %q: job %q not found", step.Name, step.Job)
		}

		// Merge envs: CLI > step > plan > defaults
		mergedEnv := make(map[string]string)

		// Start with plan-level env
		for k, v := range plan.Env {
			mergedEnv[k] = v
		}

		// Step env overrides plan
		for k, v := range step.Env {
			mergedEnv[k] = v
		}

		// CLI overrides everything
		for k, v := range cliEnv {
			mergedEnv[k] = v
		}

		// Validate against job contract
		if err := ValidateEnvContract(&job, mergedEnv); err != nil {
			return fmt.Errorf("step %d (%s): %w", i, step.Name, err)
		}
	}

	return nil
}
