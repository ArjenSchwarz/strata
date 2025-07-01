#\!/bin/bash
# Extract just the JSON test from integration_test.sh
export INPUT_PLAN_FILE="samples/websample.json"
export INPUT_OUTPUT_FORMAT="json"
export INPUT_COMMENT_ON_PR="false"
export GITHUB_STEP_SUMMARY="/tmp/step_summary_json.md"
export GITHUB_OUTPUT="/tmp/github_output_json.txt"

# Clean up previous runs
rm -f "$GITHUB_STEP_SUMMARY" "$GITHUB_OUTPUT"
touch "$GITHUB_STEP_SUMMARY" "$GITHUB_OUTPUT"

# Run the action
if ./action.sh > /tmp/action_output_json.log 2>&1; then
    echo "Action executed successfully"
    
    # Validate JSON output specifically
    if [ -f "$GITHUB_OUTPUT" ]; then
        json_summary=$(grep "json-summary=" "$GITHUB_OUTPUT" | cut -d"=" -f2-)
        if echo "$json_summary" | jq . >/dev/null 2>&1; then
            echo "JSON output is valid"
        else
            echo "JSON output is INVALID"
            echo "JSON content: $json_summary"
        fi
    else
        echo "GitHub output file not created"
    fi
else
    echo "Action execution failed"
fi
