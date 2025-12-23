#!/bin/bash
set -e

TOOL_NAME="go-jsonschema"
TOOL_PATH="$(go env GOPATH)/bin/${TOOL_NAME}"

# Install tool if missing
if [ ! -f "$TOOL_PATH" ]; then
    echo "Installing ${TOOL_NAME}..."
    go install github.com/atombender/go-jsonschema/...@latest
fi

# Helper to run go-jsonschema
function generate() {
    local schema_file=$1
    local output_file=$2
    local pkg_name=$3
    local extra_args=$4

    echo "Generating ${output_file} from ${schema_file}..."
    $TOOL_PATH \
        -p ${pkg_name} \
        -t \
        --tags json \
        -o ${output_file} \
        $extra_args \
        ${schema_file}
}

# Helper for portable sed in-place replacement
function replace_in_file() {
    local search=$1
    local replace=$2
    local file=$3
    # Use temporary file to avoid BSD/GNU sed differences
    sed "s/${search}/${replace}/g" "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"
}

mkdir -p internal/platform/schema

# Generate schemas

# Interaction Event
generate "specifications/schemas/interaction-event.schema.json" \
         "internal/platform/schema/interaction_event_gen.go" \
         "schema" \
         "--minimal-names"

# Security Alert
generate "specifications/schemas/security-alert.schema.json" \
         "internal/platform/schema/security_alert_gen.go" \
         "schema" \
         "--minimal-names"

# Vertex Batch Input
# Do NOT use --minimal-names to avoid collision on "Content"
generate "specifications/schemas/vertex-batch-input.schema.json" \
         "internal/platform/schema/vertex_batch_input_gen.go" \
         "schema" \
         "--struct-name-from-title=false"

# Simplify Vertex Batch Input names
replace_in_file 'VertexBatchInputSchemaJsonRequestContentsElemPartsElem' 'VertexPart' internal/platform/schema/vertex_batch_input_gen.go
replace_in_file 'VertexBatchInputSchemaJsonRequestContentsElem' 'VertexContent' internal/platform/schema/vertex_batch_input_gen.go
replace_in_file 'VertexBatchInputSchemaJsonRequestGenerationConfig' 'VertexGenerationConfig' internal/platform/schema/vertex_batch_input_gen.go
replace_in_file 'VertexBatchInputSchemaJsonRequest' 'VertexBatchRequest' internal/platform/schema/vertex_batch_input_gen.go
replace_in_file 'VertexBatchInputSchemaJson' 'VertexBatchInput' internal/platform/schema/vertex_batch_input_gen.go

# Vertex Batch Output
# Do NOT use --minimal-names
generate "specifications/schemas/vertex-batch-output.schema.json" \
         "internal/platform/schema/vertex_batch_output_gen.go" \
         "schema"

# Simplify Vertex Batch Output names
replace_in_file 'VertexAIBatchOutputPredictionCandidatesElemContentPartsElem' 'VertexOutputPart' internal/platform/schema/vertex_batch_output_gen.go
replace_in_file 'VertexAIBatchOutputPredictionCandidatesElemContent' 'VertexOutputContent' internal/platform/schema/vertex_batch_output_gen.go
replace_in_file 'VertexAIBatchOutputPredictionCandidatesElem' 'VertexOutputCandidate' internal/platform/schema/vertex_batch_output_gen.go
replace_in_file 'VertexAIBatchOutputPrediction' 'VertexOutputPrediction' internal/platform/schema/vertex_batch_output_gen.go
# Rename root to VertexBatchOutput if it is VertexAIBatchOutput
replace_in_file 'VertexAIBatchOutput' 'VertexBatchOutput' internal/platform/schema/vertex_batch_output_gen.go

# Batch Result Event
generate "specifications/schemas/batch-result-event.schema.json" \
         "internal/platform/schema/batch_result_event_gen.go" \
         "schema" \
         "--minimal-names"

echo "Generation complete."
