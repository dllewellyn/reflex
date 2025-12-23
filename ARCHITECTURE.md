# Application Architecture

The following diagram describes the high-level architecture and data flow of the Reflex application, including the ingestion pipeline, archival process, batch analysis, and dataset management.

```mermaid
graph TD
    %% Actors
    User([User / Client])
    Admin([Administrator])

    %% External Systems & Components
    subgraph "External / Tenant"
        EdgeGW["Edge Gateway (TypeScript)"]
    end

    %% Core Services
    subgraph "Ingestion & Real-time Analysis"
        Ingestor["Ingestor Service (Go)"]
        Pinecone[("Pinecone Vector DB")]
        KafkaRaw["Kafka: raw-interactions"]
    end

    %% Archival Pipeline
    subgraph "Archival (Hourly)"
        SchedulerHourly["Cloud Scheduler (Hourly)"]
        Loader["Loader Job (Go)"]
        GCS[("Google Cloud Storage")]
    end

    %% Analysis Pipeline
    subgraph "Batch Analysis (Daily)"
        SchedulerDaily["Cloud Scheduler (Daily)"]
        BatchJob["Batch Analyzer Job (Go)"]
        VertexAI["Vertex AI Batch Prediction"]
        KafkaAlerts["Kafka: security-alerts"]
    end

    %% Dataset Management
    subgraph "Dataset Management"
        DatasetLoader["Dataset Loader Service (Go)"]
        HuggingFace["HuggingFace Datasets"]
    end

    %% Flows
    %% Real-time Flow
    User -- "HTTP Request" --> EdgeGW
    EdgeGW -- "POST /ingest (Async)" --> Ingestor
    EdgeGW -- "POST /analyze (Sync)" --> Ingestor

    %% Ingestor Logic
    Ingestor -- "Produce Event" --> KafkaRaw
    Ingestor -- "Search Vectors (Sync Check)" --> Pinecone

    %% Archival Flow
    SchedulerHourly -- "Trigger" --> Loader
    Loader -- "Consume" --> KafkaRaw
    Loader -- "Write JSONL" --> GCS

    %% Batch Analysis Flow
    SchedulerDaily -- "Trigger" --> BatchJob
    BatchJob -- "Read Raw Data (Day)" --> GCS
    BatchJob -- "Submit Job" --> VertexAI
    VertexAI -- "Return Results" --> BatchJob
    BatchJob -- "Produce Alerts" --> KafkaAlerts

    %% Dataset Flow
    Admin -- "Run" --> DatasetLoader
    HuggingFace -- "Download Parquet" --> DatasetLoader
    DatasetLoader -- "Upsert Vectors" --> Pinecone

    %% Styling
    classDef service fill:#e1f5fe,stroke:#01579b,stroke-width:2px;
    classDef storage fill:#fff3e0,stroke:#e65100,stroke-width:2px;
    classDef external fill:#f3e5f5,stroke:#4a148c,stroke-width:2px;

    class Ingestor,Loader,BatchJob,DatasetLoader service;
    class Pinecone,KafkaRaw,GCS,VertexAI,KafkaAlerts,HuggingFace storage;
    class EdgeGW external;
```
