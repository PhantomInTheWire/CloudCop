# CloudCop

AI-powered cloud security tool that scans AWS environments for common misconfigurations and vulnerabilities. It provides actionable insights and even command-level fixes using LLMs to enhance cloud security with minimal setup.

## Features

- **AI-Powered Security Platform**: Built with Go (Gin) and Python (FastAPI) backends, and React.js/Next.js frontend
- **AWS Service Scanners**: Comprehensive security analysis for IAM, EC2, S3, Lambda, DynamoDB, and ECS
- **LLM-Based Remediation**: Automated command-line fixes for misconfigurations using DSPy for robust structures on unreliable LLM outputs
- **GraphQL API**: Real-time security scanning and vulnerability management with PostgreSQL database
- **Docker Deployment**: Multi-environment support (development/production)
- **Extensible Architecture**: Supports multiple cloud providers with role-based access control
- **SaaS Functionality**: Authentication middleware and Stripe payment webhook integration

## Quick Start

```bash
cd infra
docker compose up
```

Services will be available at:
- Frontend: http://localhost:3000
- API: http://localhost:8080
- AI: http://localhost:5001
- Worker: Logs in Docker Compose output

## Support

For support, please open an issue in the standard GitHub issue tracker.

## License

This project is licensed under the MIT License - see the LICENSE file for details.