"""Summarization service implementation using gRPC with OpenAI LLM integration."""

import json
import logging
import os
from collections import defaultdict
from concurrent import futures
from typing import Any

import grpc
from openai import OpenAI

from app.grpc_gen import summarization_pb2, summarization_pb2_grpc

logger = logging.getLogger(__name__)


class LLMClient:
    """OpenAI-compatible LLM client for summarization."""

    def __init__(self) -> None:
        self.api_key = os.getenv("OPENAI_API_KEY", "")
        self.base_url = os.getenv("OPENAI_BASE_URL", "https://openrouter.ai/api/v1")
        self.model = os.getenv("OPENAI_MODEL", "z-ai/glm-4.5-air:free")

        if not self.api_key:
            logger.warning("OPENAI_API_KEY not set, LLM features will be disabled")
            self.client = None
        else:
            self.client = OpenAI(api_key=self.api_key, base_url=self.base_url)
            logger.info(f"LLM client initialized with model: {self.model}")

    def summarize_issues(
        self,
        service: str,
        region: str,
        account_id: str,
        findings: list[str],
    ) -> tuple[str, str]:
        """Generate a summary and remedy description for findings.

        Returns:
            Tuple of (summary, remedy)
        """
        if not self.client:
            return self._fallback_summary(findings), self._fallback_remedy(service)

        max_findings = 50
        findings_text = "\n".join(findings[:max_findings])

        system_prompt = (
            f"You are a cloud security expert analyzing AWS findings. "
            f"You will analyze findings from service {service} in region {region} "
            f"for AWS account {account_id}. "
            f"Provide concise, actionable security analysis."
        )

        user_prompt = f"""Analyze these security findings and provide:
1. A brief summary of the problems found (2-3 sentences)
2. A general description of remediation steps (2-3 sentences)

Findings:
{findings_text}

Respond in JSON format:
{{"summary": "...", "remedy": "..."}}"""

        try:
            response = self.client.chat.completions.create(
                model=self.model,
                messages=[
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": user_prompt},
                ],
                temperature=0.3,
                max_tokens=500,
            )

            content = response.choices[0].message.content or "{}"
            # Try to parse JSON from response
            try:
                # Handle markdown code blocks
                if "```json" in content:
                    content = content.split("```json")[1].split("```")[0]
                elif "```" in content:
                    content = content.split("```")[1].split("```")[0]

                result = json.loads(content.strip())
                return result.get("summary", ""), result.get("remedy", "")
            except json.JSONDecodeError:
                logger.warning(f"Failed to parse LLM response as JSON: {content}")
                return content, ""

        except Exception as e:
            logger.error(f"LLM summarization failed: {e}")
            return self._fallback_summary(findings), self._fallback_remedy(service)

    def generate_commands(
        self,
        service: str,
        region: str,
        account_id: str,
        summary: str,
        remedy: str,
        resource_ids: list[str],
    ) -> list[str]:
        """Generate AWS CLI commands for remediation.

        Returns:
            List of AWS CLI commands
        """
        if not self.client:
            return self._fallback_commands(service, resource_ids)

        system_prompt = (
            f"You are an AWS automation expert. "
            f"Generate AWS CLI commands to remediate security issues in {service} "
            f"for account {account_id} in region {region}. "
            f"Only provide valid, executable AWS CLI commands."
        )

        user_prompt = f"""Generate AWS CLI commands for remediation:

Summary: {summary}

Remedy: {remedy}

Affected resources: {", ".join(resource_ids[:10])}

Respond with a JSON array of AWS CLI commands:
{{"commands": ["aws ...", "aws ..."]}}

Important:
- Use the correct region: {region}
- Commands should be safe and follow best practices
- Include comments as separate strings if needed"""

        try:
            response = self.client.chat.completions.create(
                model=self.model,
                messages=[
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": user_prompt},
                ],
                temperature=0.2,
                max_tokens=1000,
            )

            content = response.choices[0].message.content or "{}"
            try:
                # Handle markdown code blocks
                if "```json" in content:
                    content = content.split("```json")[1].split("```")[0]
                elif "```" in content:
                    content = content.split("```")[1].split("```")[0]

                result = json.loads(content.strip())
                commands = result.get("commands", [])
                # Ensure all commands are strings
                return [str(cmd) for cmd in commands if cmd]
            except json.JSONDecodeError:
                logger.warning(f"Failed to parse commands JSON: {content}")
                return []

        except Exception as e:
            logger.error(f"LLM command generation failed: {e}")
            return self._fallback_commands(service, resource_ids)

    def _fallback_summary(self, findings: list[str]) -> str:
        """Generate fallback summary without LLM."""
        count = len(findings)
        return f"Found {count} security issues that require attention."

    def _fallback_remedy(self, service: str) -> str:
        """Generate fallback remedy without LLM."""
        return f"Review and remediate {service.upper()} security configurations."

    def _fallback_commands(self, service: str, resource_ids: list[str]) -> list[str]:
        """Generate fallback commands without LLM."""
        commands = []
        if service == "s3":
            for rid in resource_ids[:3]:
                commands.append(
                    f"# Enable encryption for bucket {rid}\n"
                    f"aws s3api put-bucket-encryption --bucket {rid} "
                    "--server-side-encryption-configuration "
                    '\'{"Rules":[{"ApplyServerSideEncryptionByDefault":'
                    '{"SSEAlgorithm":"AES256"}}]}\''
                )
        elif service == "ec2":
            commands.append(
                "# Review security group rules\n"
                "aws ec2 describe-security-groups --query "
                "'SecurityGroups[?IpPermissions[?IpRanges[?CidrIp==`0.0.0.0/0`]]]'"
            )
        return commands


class SummarizationServicer(summarization_pb2_grpc.SummarizationServiceServicer):
    """Implementation of the SummarizationService gRPC service."""

    def __init__(self) -> None:
        self.llm = LLMClient()

    def SummarizeFindings(
        self,
        request: summarization_pb2.SummarizeFindingsRequest,
        context: grpc.ServicerContext,
    ) -> summarization_pb2.SummarizeFindingsResponse:
        """Summarize and group security findings with LLM-powered analysis."""
        findings = list(request.findings)
        account_id = request.account_id or "unknown"
        include_remediation = (
            request.options.include_remediation if request.options else True
        )

        logger.info(f"Summarizing {len(findings)} findings for account {account_id}")

        # Group findings by check_id and service
        grouped = self._group_findings(findings)

        # Create finding groups with LLM summaries
        finding_groups = []
        action_items = []

        total_groups = len(grouped)
        logger.info(f"Processing {total_groups} finding groups...")

        for i, (group_key, group_findings) in enumerate(grouped.items(), 1):
            logger.info(f"Processing group {i}/{total_groups}: {group_key}")

            group = self._create_finding_group(group_key, group_findings, account_id)
            finding_groups.append(group)

            # Generate action items with remediation commands for failed findings
            failed_findings = [
                f
                for f in group_findings
                if f.status == summarization_pb2.FINDING_STATUS_FAIL
            ]
            if failed_findings and include_remediation:
                logger.info(f"Generating remediation for group {i}/{total_groups}...")
                action = self._create_action_item(group, account_id, failed_findings)
                action_items.append(action)

        logger.info("Finished processing all groups.")

        # Calculate risk summary
        risk_summary = self._calculate_risk_summary(findings)

        return summarization_pb2.SummarizeFindingsResponse(
            scan_id=request.scan_id,
            groups=finding_groups,
            risk_summary=risk_summary,
            action_items=action_items,
        )

    def StreamSummarizeFindings(
        self,
        request_iterator: Any,
        context: grpc.ServicerContext,
    ) -> summarization_pb2.SummarizeFindingsResponse:
        """Stream findings and return summarized results."""
        findings = list(request_iterator)
        mock_request = summarization_pb2.SummarizeFindingsRequest(
            scan_id="streaming",
            account_id="unknown",
            findings=findings,
        )
        return self.SummarizeFindings(mock_request, context)

    def _group_findings(
        self, findings: list[summarization_pb2.Finding]
    ) -> dict[str, list[summarization_pb2.Finding]]:
        """Group findings by check_id and service."""
        grouped: dict[str, list[summarization_pb2.Finding]] = defaultdict(list)
        for finding in findings:
            key = f"{finding.service}:{finding.check_id}"
            grouped[key].append(finding)
        return grouped

    def _create_finding_group(
        self,
        group_key: str,
        findings: list[summarization_pb2.Finding],
        account_id: str,
    ) -> summarization_pb2.FindingGroup:
        """Create a FindingGroup with LLM-generated summary."""
        if not findings:
            return summarization_pb2.FindingGroup()

        first = findings[0]
        service, check_id = group_key.split(":", 1)
        region = first.region or "us-east-1"

        # Count failed findings
        failed_findings = [
            f for f in findings if f.status == summarization_pb2.FINDING_STATUS_FAIL
        ]
        failed_count = len(failed_findings)

        # Determine highest severity
        severities = [f.severity for f in findings]
        max_severity = max(severities) if severities else 0

        # Collect resource IDs
        resource_ids = [f.resource_id for f in findings]

        # Generate title
        if failed_count > 0:
            title = f"{failed_count} {service.upper()} resources failed {check_id}"
        else:
            title = f"All {len(findings)} {service.upper()} resources passed {check_id}"

        # Calculate risk score
        risk_score = self._calculate_group_risk_score(findings, max_severity)

        # Determine recommended action
        action = self._determine_action(max_severity, failed_count)

        # Collect compliance frameworks
        compliance = set()
        for f in findings:
            compliance.update(f.compliance)

        # Generate LLM summary and remedy for failed findings
        summary = ""
        remedy = ""
        if failed_count > 0:
            finding_texts = [
                f"{f.title}: {f.description}" for f in failed_findings[:20]
            ]
            summary, remedy = self.llm.summarize_issues(
                service, region, account_id, finding_texts
            )

        return summarization_pb2.FindingGroup(
            group_id=group_key,
            title=title,
            description=first.description,
            severity=max_severity,
            finding_count=len(findings),
            resource_ids=resource_ids,
            check_id=check_id,
            service=service,
            compliance=list(compliance),
            risk_score=risk_score,
            recommended_action=action,
            summary=summary,
            remedy=remedy,
        )

    def _calculate_group_risk_score(
        self,
        findings: list[summarization_pb2.Finding],
        max_severity: int,
    ) -> int:
        """Calculate risk score for a group (0-100)."""
        severity_scores = {
            summarization_pb2.SEVERITY_LOW: 25,
            summarization_pb2.SEVERITY_MEDIUM: 50,
            summarization_pb2.SEVERITY_HIGH: 75,
            summarization_pb2.SEVERITY_CRITICAL: 100,
        }
        base_score = severity_scores.get(max_severity, 0)

        failed_count = len(
            [f for f in findings if f.status == summarization_pb2.FINDING_STATUS_FAIL]
        )

        if failed_count == 0:
            return 0

        scale_factor = min(1.0 + (failed_count - 1) * 0.1, 1.5)
        return min(int(base_score * scale_factor), 100)

    def _determine_action(
        self, max_severity: int, failed_count: int
    ) -> summarization_pb2.ActionType:
        """Determine recommended action based on severity and count."""
        if failed_count == 0:
            return summarization_pb2.ACTION_TYPE_UNSPECIFIED

        if max_severity == summarization_pb2.SEVERITY_CRITICAL:
            return summarization_pb2.ACTION_TYPE_ESCALATE
        elif max_severity == summarization_pb2.SEVERITY_HIGH:
            return summarization_pb2.ACTION_TYPE_ALERT
        else:
            return summarization_pb2.ACTION_TYPE_SUGGEST_FIX

    def _create_action_item(
        self,
        group: summarization_pb2.FindingGroup,
        account_id: str,
        failed_findings: list[summarization_pb2.Finding],
    ) -> summarization_pb2.ActionItem:
        """Create an action item with LLM-generated CLI commands."""
        region = failed_findings[0].region if failed_findings else "us-east-1"

        # Generate remediation commands using LLM
        commands = self.llm.generate_commands(
            service=group.service,
            region=region,
            account_id=account_id,
            summary=group.summary,
            remedy=group.remedy,
            resource_ids=list(group.resource_ids),
        )

        return summarization_pb2.ActionItem(
            action_id=f"action_{group.group_id}",
            action_type=group.recommended_action,
            severity=group.severity,
            title=f"Fix: {group.title}",
            description=f"Address {group.finding_count} findings for {group.check_id}",
            group_id=group.group_id,
            commands=commands,
        )

    def _calculate_risk_summary(
        self, findings: list[summarization_pb2.Finding]
    ) -> summarization_pb2.RiskSummary:
        """Calculate overall risk summary."""
        critical_count = 0
        high_count = 0
        medium_count = 0
        low_count = 0
        passed_count = 0

        for f in findings:
            if f.status == summarization_pb2.FINDING_STATUS_PASS:
                passed_count += 1
            elif f.status == summarization_pb2.FINDING_STATUS_FAIL:
                if f.severity == summarization_pb2.SEVERITY_CRITICAL:
                    critical_count += 1
                elif f.severity == summarization_pb2.SEVERITY_HIGH:
                    high_count += 1
                elif f.severity == summarization_pb2.SEVERITY_MEDIUM:
                    medium_count += 1
                else:
                    low_count += 1

        total_failed = critical_count + high_count + medium_count + low_count
        if total_failed == 0:
            overall_score = 0
            risk_level = "LOW"
        else:
            weighted = (
                critical_count * 100
                + high_count * 75
                + medium_count * 50
                + low_count * 25
            )
            overall_score = min(weighted // max(total_failed, 1), 100)

            if critical_count > 0:
                risk_level = "CRITICAL"
            elif high_count > 0:
                risk_level = "HIGH"
            elif medium_count > 0:
                risk_level = "MEDIUM"
            else:
                risk_level = "LOW"

        summary_text = self._generate_summary_text(
            critical_count, high_count, medium_count, low_count, passed_count
        )

        return summarization_pb2.RiskSummary(
            overall_score=overall_score,
            critical_count=critical_count,
            high_count=high_count,
            medium_count=medium_count,
            low_count=low_count,
            passed_count=passed_count,
            risk_level=risk_level,
            summary_text=summary_text,
        )

    def _generate_summary_text(
        self,
        critical: int,
        high: int,
        medium: int,
        low: int,
        passed: int,
    ) -> str:
        """Generate a human-readable summary."""
        total_failed = critical + high + medium + low
        total = total_failed + passed

        if total_failed == 0:
            return f"All {total} security checks passed. No issues detected."

        parts = []
        if critical > 0:
            parts.append(f"{critical} critical")
        if high > 0:
            parts.append(f"{high} high")
        if medium > 0:
            parts.append(f"{medium} medium")
        if low > 0:
            parts.append(f"{low} low")

        issues_str = ", ".join(parts)
        return (
            f"Found {total_failed} security issues ({issues_str} severity) "
            f"out of {total} total checks. {passed} checks passed."
        )


def serve(port: int = 50051) -> grpc.Server:
    """Start the gRPC server."""
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    summarization_pb2_grpc.add_SummarizationServiceServicer_to_server(
        SummarizationServicer(), server
    )
    server.add_insecure_port(f"[::]:{port}")
    return server


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)

    grpc_server = serve()
    grpc_server.start()
    logger.info("Summarization gRPC server started on port 50051")
    grpc_server.wait_for_termination()
