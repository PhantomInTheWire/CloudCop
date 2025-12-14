// Package ecs provides ECS security scanning capabilities.
package ecs

import (
	"context"
	"fmt"
	"strings"

	"cloudcop/api/internal/scanner"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

var sensitiveEnvPatterns = []string{
	"SECRET", "PASSWORD", "KEY", "TOKEN", "CREDENTIAL", "API_KEY", "PRIVATE", "AUTH",
}

func (e *Scanner) checkPrivilegedContainers(_ context.Context, taskDef *types.TaskDefinition) []scanner.Finding {
	taskDefArn := aws.ToString(taskDef.TaskDefinitionArn)
	var findings []scanner.Finding

	for _, container := range taskDef.ContainerDefinitions {
		containerName := aws.ToString(container.Name)
		if container.Privileged != nil && *container.Privileged {
			findings = append(findings, e.createFinding(
				"ecs_privileged_container",
				taskDefArn,
				"ECS task has privileged container",
				fmt.Sprintf("Container %s in task %s runs in privileged mode", containerName, taskDefArn),
				scanner.StatusFail,
				scanner.SeverityCritical,
			))
		}
	}

	if len(findings) == 0 {
		findings = append(findings, e.createFinding(
			"ecs_privileged_container",
			taskDefArn,
			"ECS task has no privileged containers",
			fmt.Sprintf("Task %s has no privileged containers", taskDefArn),
			scanner.StatusPass,
			scanner.SeverityCritical,
		))
	}
	return findings
}

func (e *Scanner) checkPublicRegistry(_ context.Context, taskDef *types.TaskDefinition) []scanner.Finding {
	taskDefArn := aws.ToString(taskDef.TaskDefinitionArn)
	var findings []scanner.Finding

	publicRegistries := []string{"docker.io", "gcr.io", "ghcr.io", "quay.io", "registry.hub.docker.com"}
	for _, container := range taskDef.ContainerDefinitions {
		image := aws.ToString(container.Image)
		for _, registry := range publicRegistries {
			if strings.Contains(image, registry) || !strings.Contains(image, ".") {
				findings = append(findings, e.createFinding(
					"ecs_public_registry",
					taskDefArn,
					"ECS container uses public registry",
					fmt.Sprintf("Container image %s is from a public registry", image),
					scanner.StatusFail,
					scanner.SeverityMedium,
				))
				break
			}
		}
	}
	return findings
}

func (e *Scanner) checkTaskIAMRole(_ context.Context, taskDef *types.TaskDefinition) []scanner.Finding {
	taskDefArn := aws.ToString(taskDef.TaskDefinitionArn)
	if taskDef.TaskRoleArn != nil && *taskDef.TaskRoleArn != "" {
		return []scanner.Finding{e.createFinding(
			"ecs_task_iam_role",
			taskDefArn,
			"ECS task has IAM role assigned",
			fmt.Sprintf("Task %s has task role %s", taskDefArn, aws.ToString(taskDef.TaskRoleArn)),
			scanner.StatusPass,
			scanner.SeverityMedium,
		)}
	}
	return []scanner.Finding{e.createFinding(
		"ecs_task_iam_role",
		taskDefArn,
		"ECS task has no IAM role",
		fmt.Sprintf("Task %s has no task IAM role assigned", taskDefArn),
		scanner.StatusFail,
		scanner.SeverityMedium,
	)}
}

func (e *Scanner) checkNetworkMode(_ context.Context, taskDef *types.TaskDefinition) []scanner.Finding {
	taskDefArn := aws.ToString(taskDef.TaskDefinitionArn)
	if taskDef.NetworkMode == types.NetworkModeAwsvpc {
		return []scanner.Finding{e.createFinding(
			"ecs_awsvpc_mode",
			taskDefArn,
			"ECS task uses awsvpc network mode",
			fmt.Sprintf("Task %s uses awsvpc network mode", taskDefArn),
			scanner.StatusPass,
			scanner.SeverityMedium,
		)}
	}
	return []scanner.Finding{e.createFinding(
		"ecs_awsvpc_mode",
		taskDefArn,
		"ECS task does not use awsvpc network mode",
		fmt.Sprintf("Task %s uses %s network mode (awsvpc recommended)", taskDefArn, taskDef.NetworkMode),
		scanner.StatusFail,
		scanner.SeverityMedium,
	)}
}

func (e *Scanner) checkSecretsInEnv(_ context.Context, taskDef *types.TaskDefinition) []scanner.Finding {
	taskDefArn := aws.ToString(taskDef.TaskDefinitionArn)
	var findings []scanner.Finding

	for _, container := range taskDef.ContainerDefinitions {
		containerName := aws.ToString(container.Name)
		for _, env := range container.Environment {
			envName := aws.ToString(env.Name)
			upperName := strings.ToUpper(envName)
			for _, pattern := range sensitiveEnvPatterns {
				if strings.Contains(upperName, pattern) {
					findings = append(findings, e.createFinding(
						"ecs_secrets_in_env",
						taskDefArn,
						"ECS container has secrets in environment variables",
						fmt.Sprintf("Container %s has sensitive env var %s (use secrets)", containerName, envName),
						scanner.StatusFail,
						scanner.SeverityHigh,
					))
					break
				}
			}
		}
	}
	return findings
}

func (e *Scanner) checkCloudWatchLogs(_ context.Context, taskDef *types.TaskDefinition) []scanner.Finding {
	taskDefArn := aws.ToString(taskDef.TaskDefinitionArn)
	var findings []scanner.Finding

	for _, container := range taskDef.ContainerDefinitions {
		containerName := aws.ToString(container.Name)
		hasLogs := false
		if container.LogConfiguration != nil && container.LogConfiguration.LogDriver == types.LogDriverAwslogs {
			hasLogs = true
		}
		if hasLogs {
			findings = append(findings, e.createFinding(
				"ecs_cloudwatch_logs",
				taskDefArn,
				"ECS container has CloudWatch Logs configured",
				fmt.Sprintf("Container %s logs to CloudWatch", containerName),
				scanner.StatusPass,
				scanner.SeverityMedium,
			))
		} else {
			findings = append(findings, e.createFinding(
				"ecs_cloudwatch_logs",
				taskDefArn,
				"ECS container has no CloudWatch Logs",
				fmt.Sprintf("Container %s does not log to CloudWatch", containerName),
				scanner.StatusFail,
				scanner.SeverityMedium,
			))
		}
	}
	return findings
}
