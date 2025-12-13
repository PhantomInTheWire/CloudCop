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

### Prerequisites

- Docker 24+
- Docker Compose v2+
- 8GB RAM minimum

### Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/PhantomInTheWire/CloudCop.git
   cd CloudCop
   ```

2. Copy environment file:
   ```bash
   cp .env.example .env
   # Edit .env and add your OPENAI_API_KEY
   ```

3. Start all services:
   ```bash
   npm run dev
   ```

Services will be available at:
- **Frontend**: http://localhost:3000 - Next.js Dashboard
- **Backend GraphQL**: http://localhost:8080/graphql - Go API
- **Kestra**: http://localhost:8081 - AI Agent Orchestration
- **Neo4j Browser**: http://localhost:7474 - Security Graph (user: neo4j, password: from .env)
- **PostgreSQL**: localhost:5432 - Metadata & Results

## Support


For support, please open an issue in the standard GitHub issue tracker.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
