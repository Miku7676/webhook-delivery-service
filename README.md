# Webhook Delivery System - Biswajit Bal

This project implements a **Webhook Delivery System** that enables:

-   **Subscription Management**: Create, retrieve, update, and delete subscriptions with optional secret keys for signature verification.
    
-   **Webhook Ingestion and Queuing**: Accept incoming webhook payloads and enqueue them asynchronously using **Redis** and **Asynq**.

-   **Reliable Delivery with Retry and Backoff**: Deliver webhook events with an automatic **exponential retry strategy** on failure.
    
-   **Payload Signature Verification**: Authenticate incoming webhook payloads using **HMAC-SHA256**, ensuring secure transmission.
    
-   **Delivery Logging and Monitoring**: Track webhook delivery status, retries, and failures with detailed logs stored in **PostgreSQL**.
- **Background Service And cleanups**: Queued are processed by background services that also cleanup logs older than 24 hours
    
-   **Production-Ready Deployment**: Fully **Dockerized** with a multi-stage build, **Swagger/OpenAPI** documentation, and auto-deployment on **Render**.

✅ Deployed at: [`https://webhook-api-wwhi.onrender.com`](https://webhook-api-wwhi.onrender.com)  
✅ Swagger UI: [`https://webhook-api-wwhi.onrender.com/swagger/index.html`](https://webhook-api-wwhi.onrender.com/swagger/index.html)

# Index
- [Technical Specifications](#technical-specifications)
- [Deployment Architecture](#deployment-architecture)
- [API Endpoints](#api-endpoints)
- [Setup Instructions](#setup-instructions)
- [Payload Signature Verification](#payload-signature-verification)
- [Performance Strategy](#performance-strategy)
- [Important Notes And Assumptions](#important-notes-and-assumptions)
- [Monthly Cost Estimation](#monthly-cost-estimation)
- [Libraries and Tools](#libraries-and-tools)


#  Technical Specifications

-   **Programming Language**: Go (Golang)
-   **Framework**: Gin Web Framework
-   **Asynchronous**: Asynq
-   **Database**:  PostgreSQL
-   **Caching**:  Redis
-   **Payload Verification**:  HMAC-SHA256 signature verification (based on a subscription-specific secret key.)
-   **Containerization**: Docker
-   **API Documentation and UI**: Swagger/OpenAPI 3.0
-   **Deployment Platform**: Render.com (Free Tier)


# Deployment Architecture

```plaintext
User ----> API Server (Render) ----> Postgres (Render)
                     |                  
                     V
              Redis (Upstash)
                     |
                     V
           Worker Server (Render)
                     |
                     V
              Target URL (Webhook Delivery)
```


# API Endpoints


### 1. Create a New Subscription

**Endpoint:**  
`POST /subscriptions`

**Sample CURL:**
```bash
curl -X POST https://webhook-api-wwhi.onrender.com/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "target_url": "https://webhook.site/your-webhook-id",
    "secret": "mysecretkey"
}'
```

**Expected Response:**

```json
{  
	"id":  "90c80e5d-aaa5-4651-9da3-ba4cafee5a7a",  
	"target_url":  "https://webhook.site/your-webhook-id",  
	"secret":  "mysecretkey"  
}
```

**HTTP Status:** 201 Created


### 2. Get a Subscription by ID

**Endpoint:**  
`GET /subscriptions/{id}`

**Sample CURL:**

```bash
curl -X GET https://webhook-api-wwhi.onrender.com/subscriptions/{subscription_id}
```
**Expected Response:**

```json
{  
	"id":  "90c80e5d-aaa5-4651-9da3-ba4cafee5a7a",  
	"target_url":  "https://webhook.site/your-webhook-id",  
	"secret":  "mysecretkey"  
}
```
**HTTP Status:** 200 OK


## 3. Update a Subscription

**Endpoint:**  
`PUT /subscriptions/{id}`

**Sample CURL:**

```bash
curl -X PUT https://webhook-api-wwhi.onrender.com/subscriptions/{subscription_id} \
  -H "Content-Type: application/json" \
  -d '{
    "target_url": "https://webhook.site/updated-webhook-id",
    "secret": "updatedsecretkey"
}'
```
**Expected Response:**

```json
{  
	"id":  "90c80e5d-aaa5-4651-9da3-ba4cafee5a7a",
	"target_url":  "https://webhook.site/updated-webhook-id",  
	"secret":  "updatedsecretkey"  
}
```
**HTTP Status:** 200 OK


## 4. Delete a Subscription

**Endpoint:**  
`DELETE /subscriptions/{id}`

**Sample CURL:**

```bash
curl -X DELETE https://webhook-api-wwhi.onrender.com/subscriptions/{subscription_id}
```

**HTTP Status:** 204 No Content

## 5. Ingest a Webhook (Send Webhook Payload)

**Endpoint:**  
`POST /ingest/{subscription_id}`

**Sample CURL:**

```bash
curl -X POST https://webhook-api-wwhi.onrender.com/ingest/{subscription_id} \
  -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: {computed_signature}" \
  -d '{
    "event": "order.created",
    "order_id": "12345"
}'
``` 

**Expected Response:**

```json
{  
	"status":  "queued",  
	"task_id":  "f67251c2-02f8-44be-b61b-fc76e10c1d8d"  
}
```
**HTTP Status:** 202 Accepted

✅ Note: `X-Hub-Signature-256` header is **mandatory** if subscription has a secret.


## 6. Check Status of a Webhook Delivery

**Endpoint:**  
`GET /status/{webhook_id}`

**Sample CURL:**

```bash
curl -X GET https://webhook-api-wwhi.onrender.com/status/{webhook_id}
```
**Expected Response:**

```json
{  
	"id":  "f67251c2-02f8-44be-b61b-fc76e10c1d8d",  
	"subscription_id":  "90c80e5d-aaa5-4651-9da3-ba4cafee5a7a",	
	"target_url":  "https://webhook.site/your-webhook-id",  
	"attempt_number":  1,  
	"status":  "Success",  
	"http_status":  200,  
	"error_message":  "",  
	"created_at":  "2025-04-26T12:30:45Z"
}
```
**HTTP Status:** 200 OK

✅ Shows latest delivery attempt details.


## 7. Get Recent Delivery Logs for a Subscription

**Endpoint:**  
`GET /subscriptions/{id}/logs`

**Sample CURL:**

```bash
curl -X GET https://webhook-api-wwhi.onrender.com/subscriptions/{subscription_id}/logs
```

**Expected Response:**

```json
[  
	{  
		"id":  "log-entry-id-2",  
		"webhook_task_id":  "f67251c2-02f8-44be-b61b-fc76e10c1d8d",  
		"subscription_id":  "90c80e5d-aaa5-4651-9da3-ba4cafee5a7a",  
		"target_url":  "https://webhook.site/your-webhook-id",  
		"attempt_number":  2,  
		"status":  "Success",  
		"http_status":  200,  
		"error_message":  "",  
		"created_at":  "2025-04-26T12:30:45Z"  
	},  
	{
		"id":  "log-entry-id-1",  
		"webhook_task_id":  "another-task-id",  
		"subscription_id":  "90c80e5d-aaa5-4651-9da3-ba4cafee5a7a",  
		"target_url":  "https://webhook.site/your-webhook-id",  
		"attempt_number":  1,  
		"status":  "Failed",  
		"http_status":  500,  
		"error_message":  "Server Error",  
		"created_at":  "2025-04-26T12:40:12Z"  
	}  
]
```
**HTTP Status:** 200 OK

✅ Shows multiple delivery attempts and their statuses.



# Setup Instructions

### 1. Run Locally (Docker Compose)

```bash
docker-compose up --build
```
-   API Service: `http://localhost:8080`
-   Swagger UI: `http://localhost:8080/swagger/index.html`
    

----------

### 2. Deployment on Render

-   Code is pushed to GitHub (`main` branch).
    
-   Render Auto-Deploys on every Git push.
    
-   Render services:
    
    -   `webhook-api`: API server
        
    -   `webhook-worker`: Worker server
        
    -   Upstash Redis: used for background queueing
        
    -   Render-managed Postgres: for persistence
        

# Payload Signature Verification 

When creating a subscription, a **secret** can be optionally set.
-   During ingestion, the system expects an `X-Hub-Signature-256` header.
-   The header must contain the **HMAC SHA-256** hash of the payload, using the secret.
-   If the signature is invalid or missing:
    -   Webhook ingestion is rejected with **401 Unauthorized**.
        





# Performance Strategy
-   Subscription details are cached in Redis once a task is queued.
-   When delivering webhooks, the worker first checks Redis for subscription data.
-   If not found, falls back to database and caches it.
    

Redis reduces database load and speeds up retrieval.




#  Important Notes And Assumptions

- **Cold Starts**: Since using Free Tier on Render, services may sleep after inactivity, causing cold starts (~3-5 seconds) when triggered again.
- No Tests were included. Tests performed using custom curl scripts and Webhook.site.
- For the current expected traffic (5000 webhooks/day), the Free Tier is **sufficient and sustainable**.
- All payloads are assumed to be json format.
- This project does not implement custom exponential backoff rather uses builtin backoff function in **Asynq** library


# Monthly Cost Estimation



## Updated Assumptions

| Parameter | Value |
|-----------|-------|
| Webhooks ingested per day | 5000+ |
| Average delivery attempts per webhook | 1.5 |
| Total delivery attempts per day | ~7500 |
| Month duration | 30 days |
| Total delivery attempts per month | ~225,000 |
| Continuous 24x7 operation | Yes |

✅ More activity, heavier load than Free Tier.


## Service-wise Updated Estimation



### 1. Render Web Service (API Server)

| Item | Value |
|------|-------|
| Plan | Starter Plan |
| Cost | **$7/month** |
| Includes | 512MB RAM, 0.5 CPU, No auto-sleep |
| Notes | Immediate cold-start-free response, 24x7 alive |

✅ Needed because Free Web Service has auto-sleep.

---

### 2. Render Background Worker (Webhook Delivery Worker)

| Item | Value |
|------|-------|
| Plan | Starter Plan |
| Cost | **$7/month** |
| Includes | 512MB RAM, 0.5 CPU, No auto-sleep |
| Notes | Worker stays active 24x7, processes retries faster |

✅ Needed for heavier retry load, background delivery.

---

### 3. Render Managed PostgreSQL (Database)

| Item | Value |
|------|-------|
| Plan | Starter Database |
| Cost | **$7/month** |
| Includes | 1 GB storage, 1 concurrent connection |
| Notes | Enough for 100K–500K delivery logs and subscriptions |

✅ Sufficient for your workload.

---

### 4. Upstash Redis (Queue + Cache)

| Item | Value |
|------|-------|
| Plan | Pay As You Go (beyond Free Tier) |
| Free Commands | 10,000 commands/day free |
| Needed Commands | ~15,000 commands/day |
| Estimated Monthly Usage | ~450,000 commands/month |
| Cost Estimation | **~$5/month** |

✅ Upstash pricing is ~$0.20 per 100,000 additional commands.

Calculation:

```plaintext
(450,000 - 300,000 free) = 150,000 excess commands
Cost = 150,000 * $0.20 / 100,000 = ~$0.30
Rounded minimum billing ~ $5/month based on current plans
```


# Libraries and Tools
- **Gin** ([github.com/gin-gonic/gin](https://github.com/gin-gonic/gin)) — Lightweight Go web framework for building REST APIs.
- **GORM** ([gorm.io](https://gorm.io)) — ORM library for Golang to interact with PostgreSQL.
- **Asynq** ([github.com/hibiken/asynq](https://github.com/hibiken/asynq)) — Redis-based asynchronous task queue for background job processing.
- **Redis** ([redis.io](https://redis.io)) — In-memory key-value store for caching and task queuing.
- **Upstash Redis** ([upstash.com](https://upstash.com)) — Serverless managed Redis service for hosting queues and cache.
- **PostgreSQL** ([postgresql.org](https://www.postgresql.org)) — Open-source relational database for persistent data storage.
- **Swaggo/swag** ([github.com/swaggo/swag](https://github.com/swaggo/swag)) — Tool to generate Swagger/OpenAPI documentation from Go annotations.
- **Docker** ([docker.com](https://www.docker.com)) — Containerization platform to package the application and its dependencies.
- **Docker Compose** ([docs.docker.com/compose](https://docs.docker.com/compose/)) — Tool for defining and running multi-container Docker applications.
- **Render** ([render.com](https://render.com)) — Cloud hosting platform used for deploying the API and Worker with automatic GitHub CI/CD.
- **RapidAPI Extension for GitHub** ([RapidAPI Docs](https://docs.rapidapi.com/docs/github-integration)) — Extension used for API testing, management, and documentation within GitHub repositories.
- **godotenv** ([github.com/joho/godotenv](https://github.com/joho/godotenv)) — Loads environment variables from `.env` files into Go applications during local development.
- **Gin CORS Middleware** ([github.com/gin-contrib/cors](https://github.com/gin-contrib/cors)) — Middleware to handle Cross-Origin Resource Sharing (CORS) in Gin applications.
- **ChatGPT** — Used for guidance, code refinement, debugging strategies, and best practice confirmations during development.
- **Claude** — Assisted in brainstorming architecture design ideas, writing explanations, and improving documentation.

# 🙏 Thank you! 
