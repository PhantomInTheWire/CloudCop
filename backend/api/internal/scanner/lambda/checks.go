// Package lambda provides Lambda security scanning capabilities.
package lambda

import (
	"context"
	"fmt"
	"strings"

	"cloudcop/api/internal/scanner"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

var sensitiveEnvVarPatterns = []string{
	"SECRET", "PASSWORD", "KEY", "TOKEN", "CREDENTIAL", "API_KEY",
	"PRIVATE", "AUTH", "PASS", "PWD", "ACCESS",
}

const maxRecommendedTimeoutSeconds = 300 // 5 minutes

func (l *Scanner) checkEnvSecrets(_ context.Context, fn types.FunctionConfiguration) []scanner.Finding {
	fnName := aws.ToString(fn.FunctionName)
	if fn.Environment == nil || fn.Environment.Variables == nil {
		return nil
	}

	var sensitiveVars []string
	for key := range fn.Environment.Variables {
		upperKey := strings.ToUpper(key)
		for _, pattern := range sensitiveEnvVarPatterns {
			if strings.Contains(upperKey, pattern) {
				sensitiveVars = append(sensitiveVars, key)
				break
			}
		}
	}

	if len(sensitiveVars) > 0 {
		return []scanner.Finding{l.createFinding(
			"lambda_env_secrets",
			fnName,
			"Lambda function has sensitive environment variables",
			fmt.Sprintf("Function %s has potentially sensitive env vars: %v (use Secrets Manager instead)", fnName, sensitiveVars),
			scanner.StatusFail,
			scanner.SeverityCritical,
		)}
	}
	return nil
}

func (l *Scanner) checkCloudWatchLogs(_ context.Context, fn types.FunctionConfiguration) []scanner.Finding {
	fnName := aws.ToString(fn.FunctionName)
	if fn.LoggingConfig != nil && fn.LoggingConfig.LogGroup != nil {
		return []scanner.Finding{l.createFinding(
			"lambda_cloudwatch_logs",
			fnName,
			"Lambda function has CloudWatch Logs configured",
			fmt.Sprintf("Function %s logs to %s", fnName, aws.ToString(fn.LoggingConfig.LogGroup)),
			scanner.StatusPass,
			scanner.SeverityMedium,
		)}
	}
	// Lambda creates default log group automatically
	return []scanner.Finding{l.createFinding(
		"lambda_cloudwatch_logs",
		fnName,
		"Lambda function uses default CloudWatch logging",
		fmt.Sprintf("Function %s uses default log group /aws/lambda/%s", fnName, fnName),
		scanner.StatusPass,
		scanner.SeverityMedium,
	)}
}

func (l *Scanner) checkVPCConfig(_ context.Context, fn types.FunctionConfiguration) []scanner.Finding {
	fnName := aws.ToString(fn.FunctionName)
	if fn.VpcConfig != nil && len(fn.VpcConfig.SubnetIds) > 0 {
		return []scanner.Finding{l.createFinding(
			"lambda_vpc_config",
			fnName,
			"Lambda function is configured in VPC",
			fmt.Sprintf("Function %s runs in VPC with %d subnets", fnName, len(fn.VpcConfig.SubnetIds)),
			scanner.StatusPass,
			scanner.SeverityMedium,
		)}
	}
	return []scanner.Finding{l.createFinding(
		"lambda_vpc_config",
		fnName,
		"Lambda function is not in VPC",
		fmt.Sprintf("Function %s runs outside VPC (consider VPC for sensitive data access)", fnName),
		scanner.StatusFail,
		scanner.SeverityMedium,
	)}
}

func (l *Scanner) checkDLQ(_ context.Context, fn types.FunctionConfiguration) []scanner.Finding {
	fnName := aws.ToString(fn.FunctionName)
	if fn.DeadLetterConfig != nil && fn.DeadLetterConfig.TargetArn != nil {
		return []scanner.Finding{l.createFinding(
			"lambda_dlq",
			fnName,
			"Lambda function has dead letter queue configured",
			fmt.Sprintf("Function %s has DLQ: %s", fnName, aws.ToString(fn.DeadLetterConfig.TargetArn)),
			scanner.StatusPass,
			scanner.SeverityLow,
		)}
	}
	return []scanner.Finding{l.createFinding(
		"lambda_dlq",
		fnName,
		"Lambda function has no dead letter queue",
		fmt.Sprintf("Function %s has no DLQ configured (failed invocations may be lost)", fnName),
		scanner.StatusFail,
		scanner.SeverityLow,
	)}
}

func (l *Scanner) checkTracing(_ context.Context, fn types.FunctionConfiguration) []scanner.Finding {
	fnName := aws.ToString(fn.FunctionName)
	if fn.TracingConfig != nil && fn.TracingConfig.Mode == types.TracingModeActive {
		return []scanner.Finding{l.createFinding(
			"lambda_tracing",
			fnName,
			"Lambda function has X-Ray tracing enabled",
			fmt.Sprintf("Function %s has active X-Ray tracing", fnName),
			scanner.StatusPass,
			scanner.SeverityLow,
		)}
	}
	return []scanner.Finding{l.createFinding(
		"lambda_tracing",
		fnName,
		"Lambda function has X-Ray tracing disabled",
		fmt.Sprintf("Function %s does not have X-Ray tracing enabled", fnName),
		scanner.StatusFail,
		scanner.SeverityLow,
	)}
}

func (l *Scanner) checkTimeout(_ context.Context, fn types.FunctionConfiguration) []scanner.Finding {
	fnName := aws.ToString(fn.FunctionName)
	timeout := aws.ToInt32(fn.Timeout)
	if timeout > maxRecommendedTimeoutSeconds {
		return []scanner.Finding{l.createFinding(
			"lambda_timeout",
			fnName,
			"Lambda function timeout exceeds recommended limit",
			fmt.Sprintf("Function %s has timeout of %d seconds (recommended: ≤%d)", fnName, timeout, maxRecommendedTimeoutSeconds),
			scanner.StatusFail,
			scanner.SeverityLow,
		)}
	}
	return []scanner.Finding{l.createFinding(
		"lambda_timeout",
		fnName,
		"Lambda function timeout is within recommended limits",
		fmt.Sprintf("Function %s has timeout of %d seconds", fnName, timeout),
		scanner.StatusPass,
		scanner.SeverityLow,
	)}
}

func (l *Scanner) checkReservedConcurrency(ctx context.Context, fn types.FunctionConfiguration) []scanner.Finding {
	fnName := aws.ToString(fn.FunctionName)
	concurrency, err := l.client.GetFunctionConcurrency(ctx, &lambda.GetFunctionConcurrencyInput{
		FunctionName: fn.FunctionName,
	})
	if err != nil {
		// Return error finding instead of nil
		return []scanner.Finding{l.createFinding(
			"lambda_reserved_concurrency",
			fnName,
			"Could not determine reserved concurrency",
			fmt.Sprintf("Function %s: API error: %v", fnName, err),
			scanner.StatusFail,
			scanner.SeverityLow,
		)}
	}

	if concurrency.ReservedConcurrentExecutions != nil {
		return []scanner.Finding{l.createFinding(
			"lambda_reserved_concurrency",
			fnName,
			"Lambda function has reserved concurrency",
			fmt.Sprintf("Function %s has %d reserved concurrent executions", fnName, *concurrency.ReservedConcurrentExecutions),
			scanner.StatusPass,
			scanner.SeverityLow,
		)}
	}
	return []scanner.Finding{l.createFinding(
		"lambda_reserved_concurrency",
		fnName,
		"Lambda function has no reserved concurrency",
		fmt.Sprintf("Function %s uses unreserved concurrency pool", fnName),
		scanner.StatusFail,
		scanner.SeverityLow,
	)}
}
