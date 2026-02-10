# CloudCop

AI-powered cloud security tool that scans AWS environments for common misconfigurations and vulnerabilities. It provides actionable insights and even command-level fixes using LLMs to enhance cloud security with minimal setup.

## Quick Start

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
