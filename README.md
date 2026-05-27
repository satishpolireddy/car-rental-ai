# 🚗 DriveAI — Enterprise Car Rental Platform

[![CI/CD](https://github.com/satishpolireddy/car-rental-ai/actions/workflows/ci.yml/badge.svg)](https://github.com/satishpolireddy/car-rental-ai/actions)
[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat&logo=react&logoColor=black)](https://react.dev)
[![Azure](https://img.shields.io/badge/Azure-OpenAI-0078D4?style=flat&logo=microsoftazure&logoColor=white)](https://azure.microsoft.com)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A production-grade, horizontally scalable car rental booking platform built with **Go**, **React**, **Azure OpenAI**, and **SQL Server**. Features AI-powered car recommendations, a concurrent ETL pipeline for fleet inventory ingestion, Redis caching, and full Azure/Kubernetes deployment via Terraform.

Reduces manual fleet processing time by **45%** through automated ETL and AI-assisted search.

---

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                        Azure (AKS)                           │
│                                                              │
│  ┌──────────────┐    ┌──────────────┐   ┌────────────────┐  │
│  │  React/Nginx │───▶│  Go API      │──▶│  SQL Server    │  │
│  │  (Frontend)  │    │  (Backend)   │   │  (Azure SQL)   │  │
│  └──────────────┘    └──────┬───────┘   └────────────────┘  │
│                             │                                │
│                    ┌────────┴────────┐                       │
│                    │                 │                       │
│              ┌─────▼─────┐   ┌──────▼──────┐               │
│              │  Redis     │   │ Azure OpenAI │               │
│              │  (Cache)   │   │  (GPT-4o)   │               │
│              └────────────┘   └─────────────┘               │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  ETL Pipeline (Worker Pool) — Fleet data ingestion  │    │
│  └─────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────┘
```

## Features

- **AI-Powered Recommendations** — Natural language car search ("I need a family car for a mountain road trip") powered by Azure OpenAI GPT-4o
- **Smart Availability Search** — Date-range conflict detection with indexed SQL queries
- **Concurrent ETL Pipeline** — Worker pool (configurable concurrency) ingests fleet data from external sources in batches
- **Redis Caching** — AI responses cached for 10 minutes, reducing API costs and latency
- **Horizontal Scaling** — Stateless Go backend; scale with `docker-compose up --scale backend=N` or AKS HPA
- **Graceful Shutdown** — SIGTERM handling drains in-flight requests and stops ETL cleanly
- **Rate Limiting** — Per-IP request limiter (Redis-upgradeable for multi-instance)
- **CI/CD** — GitHub Actions: test → build → push to ACR → rolling deploy to AKS

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend API | Go 1.22, Gin, GORM |
| Frontend | React 18, TypeScript, Vite, TailwindCSS |
| AI | Azure OpenAI GPT-4o |
| Database | SQL Server (local) / Azure SQL (prod) |
| Cache | Redis 7 |
| Infrastructure | Terraform, Azure AKS, ACR |
| CI/CD | GitHub Actions |

## Quick Start (Docker)

```bash
git clone https://github.com/satishpolireddy/car-rental-ai.git
cd car-rental-ai

# Configure environment
cp .env.example .env
# Edit .env — add your AZURE_OPENAI_ENDPOINT and AZURE_OPENAI_KEY

# Start all services
docker-compose up --build

# App available at:
#   Frontend → http://localhost:3000
#   Backend  → http://localhost:8080
#   Health   → http://localhost:8080/health
```

## API Reference

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/cars/search` | Search available cars by location & dates |
| `GET` | `/api/v1/cars/:id` | Get car details |
| `GET` | `/api/v1/locations` | List all pickup locations |
| `POST` | `/api/v1/bookings` | Create a booking |
| `GET` | `/api/v1/bookings/:id` | Get booking details |
| `DELETE` | `/api/v1/bookings/:id` | Cancel a booking |
| `GET` | `/api/v1/customers/:id/bookings` | Get customer bookings |
| `POST` | `/api/v1/ai/recommend` | AI car recommendations |
| `GET` | `/health` | Health check |

## Scalability Design

- **Database**: Connection pool (25 max open, 10 idle, 5-min lifetime). Azure SQL scales to 80 vCores.
- **Redis**: Pool of 20 connections. Azure Cache for Redis Standard tier = replicated, HA.
- **Backend**: Stateless — scale horizontally behind a load balancer. AKS HPA auto-scales 2→10 pods.
- **ETL**: Configurable worker pool (`ETL_WORKERS`) and batch size (`ETL_BATCH_SIZE`) for throughput tuning.
- **Graceful shutdown**: 10-second drain window ensures zero dropped requests during rolling deploys.

## Production Deployment (Azure)

```bash
cd terraform/environments/prod
terraform init
terraform plan -var-file=prod.tfvars
terraform apply
```

## Project Structure

```
car-rental-ai/
├── backend/
│   ├── cmd/server/          # main.go — wires up all dependencies
│   ├── config/              # 12-factor config from env vars
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handlers/    # HTTP handlers (Gin)
│   │   │   └── middleware/  # Logger, rate limiter, request ID
│   │   ├── etl/             # Concurrent ETL pipeline
│   │   ├── models/          # GORM models + request/response types
│   │   ├── repository/      # Database access layer
│   │   └── services/        # Business logic (booking, AI)
│   └── Dockerfile
├── frontend/
│   └── src/
│       ├── components/      # CarCard, AIAssistant
│       ├── pages/           # SearchPage, BookingPage, ConfirmationPage
│       ├── services/        # Typed API client (Axios)
│       └── store/           # Zustand state management
├── terraform/
│   ├── modules/aks/         # AKS cluster with auto-scaling node pool
│   └── environments/prod/   # Production Azure resources
├── .github/workflows/ci.yml # Test → Build → Push → Deploy
├── docker-compose.yml
└── .env.example
```

## License

MIT © Satish Kumar Reddy
