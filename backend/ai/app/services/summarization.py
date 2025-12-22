"""Summarization service implementation using gRPC."""

from collections import defaultdict
from concurrent import futures
from typing import Any

import grpc

from app.grpc_gen import summarization_pb2, summarization_pb2_grpc


class SummarizationServicer(summarization_pb2_grpc.SummarizationServiceServicer):
    """Implementation of the SummarizationService gRPC service."""

    def SummarizeFindings(
        self,
        request: summarization_pb2.SummarizeFindingsRequest,
        context: grpc.ServicerContext,
    ) -> summarization_pb2.SummarizeFindingsResponse:
        """Summarize and group security findings."""
        findings = list(request.findings)

        # Group findings by check_id
        grouped = self._group_findings(findings)

        # Create finding groups
        finding_groups = []
        action_items = []

        for group_key, group_findings in grouped.items():
            group = self._create_finding_group(group_key, group_findings)
            finding_groups.append(group)

            # Generate action items for failed findings
            if any(
                f.status == summarization_pb2.FINDING_STATUS_FAIL
                for f in group_findings
            ):
                action = self._create_action_item(group)
                action_items.append(action)

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
        # Create a mock request to reuse logic
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
    ) -> summarization_pb2.FindingGroup:
        """Create a FindingGroup from a list of similar findings."""
        if not findings:
            return summarization_pb2.FindingGroup()

        first = findings[0]
        service, check_id = group_key.split(":", 1)

        # Count failed findings
        failed_count = len(
            [f for f in findings if f.status == summarization_pb2.FINDING_STATUS_FAIL]
        )

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
        )

    def _calculate_group_risk_score(
        self,
        findings: list[summarization_pb2.Finding],
        max_severity: int,
    ) -> int:
        """Calculate risk score for a group (0-100)."""
        # Base score from severity
        severity_scores = {
            summarization_pb2.SEVERITY_LOW: 25,
            summarization_pb2.SEVERITY_MEDIUM: 50,
            summarization_pb2.SEVERITY_HIGH: 75,
            summarization_pb2.SEVERITY_CRITICAL: 100,
        }
        base_score = severity_scores.get(max_severity, 0)

        # Adjust based on number of affected resources
        failed_count = len(
            [f for f in findings if f.status == summarization_pb2.FINDING_STATUS_FAIL]
        )

        if failed_count == 0:
            return 0  # No risk if all passed

        # Scale by number of failures (more failures = higher risk)
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
        self, group: summarization_pb2.FindingGroup
    ) -> summarization_pb2.ActionItem:
        """Create an action item for a finding group."""
        return summarization_pb2.ActionItem(
            action_id=f"action_{group.group_id}",
            action_type=group.recommended_action,
            severity=group.severity,
            title=f"Fix: {group.title}",
            description=f"Address {group.finding_count} findings for {group.check_id}",
            group_id=group.group_id,
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

        # Calculate overall score
        total_failed = critical_count + high_count + medium_count + low_count
        if total_failed == 0:
            overall_score = 0
            risk_level = "LOW"
        else:
            # Weighted score
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

        # Generate summary text
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
    import logging

    logging.basicConfig(level=logging.INFO)
    logger = logging.getLogger(__name__)

    grpc_server = serve()
    grpc_server.start()
    logger.info("Summarization gRPC server started on port 50051")
    grpc_server.wait_for_termination()
