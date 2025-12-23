├── functions/          # FIREBASE CLOUD FUNCTIONS
│   ├── lib/            # Compiled JavaScript (from TypeScript)
│   ├── src/            # TypeScript Source Code
│   │   └── index.ts    # Main Entrypoint
│   └── package.json    # Function Dependencies
├── internal/           # PRIVATE APPLICATION CODE (Go)
│   ├── app/            # Feature Implementations
│   │   ├── ingestor/   # Ingestor Service Logic
│   │   └── worker/     # Worker Service Logic
│   └── platform/       # Infrastructure Adapters (Kafka, Firebase, PubSub)
├── specifications/     # ARCHITECTURAL DOCUMENTATION
│   ├── tech_stack.md   # Technology Choices
│   └── directory_structure.md # (This File)
├── specs/              # FEATURE SPECIFICATIONS (Gemini CLI)
├── terraform/          # INFRASTRUCTURE AS CODE
│   └── hub/            # SHARED INFRASTRUCTURE (VPC, Kafka, Artifact Registry)
└── ...