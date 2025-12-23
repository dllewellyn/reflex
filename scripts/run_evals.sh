#!/bin/bash
set -e

# 1. Inject tests
echo "Injecting tests from test_prompts.json..."
go run scripts/inject_tests.go

# 2. Run evaluation
echo "Running evaluation..."
gh models eval prompts/security-judge.prompt.yml --json
