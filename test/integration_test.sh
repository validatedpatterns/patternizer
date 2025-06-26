#!/bin/bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test configuration
PATTERNIZER_BINARY="${PATTERNIZER_BINARY:-./src/patternizer}"
TEST_REPO_URL="https://github.com/dminnear-rh/trivial-pattern.git"
TEST_DIR="/tmp/patternizer-integration-test"
TEST_DIR_SECRETS="/tmp/patternizer-integration-test-secrets"
TEST_DIR_CUSTOM="/tmp/patternizer-integration-test-custom"

echo -e "${YELLOW}Starting patternizer integration tests...${NC}"

# Clean up any previous test runs
if [ -d "$TEST_DIR" ]; then
    rm -rf "$TEST_DIR"
fi
if [ -d "$TEST_DIR_SECRETS" ]; then
    rm -rf "$TEST_DIR_SECRETS"
fi
if [ -d "$TEST_DIR_CUSTOM" ]; then
    rm -rf "$TEST_DIR_CUSTOM"
fi

# Convert PATTERNIZER_BINARY to absolute path before changing directories
PATTERNIZER_BINARY=$(realpath "$PATTERNIZER_BINARY")

# Get the absolute path to the repository root (where resource files are located)
REPO_ROOT=$(pwd)

# Set absolute paths to expected files
EXPECTED_VALUES_GLOBAL="$REPO_ROOT/test/expected_values_global.yaml"
EXPECTED_VALUES_PROD="$REPO_ROOT/test/expected_values_prod.yaml"
EXPECTED_VALUES_PROD_WITH_SECRETS="$REPO_ROOT/test/expected_values_prod_with_secrets.yaml"
EXPECTED_VALUES_GLOBAL_CUSTOM="$REPO_ROOT/test/expected_values_global_custom.yaml"
EXPECTED_VALUES_RENAMED_CLUSTER_GROUP="$REPO_ROOT/test/expected_values_renamed_cluster_group.yaml"
INITIAL_VALUES_GLOBAL_CUSTOM="$REPO_ROOT/test/initial_values_global_custom.yaml"

# Check if patternizer binary exists and is executable
if [ ! -x "$PATTERNIZER_BINARY" ]; then
    echo -e "${RED}ERROR: Patternizer binary not found or not executable at: $PATTERNIZER_BINARY${NC}"
    exit 1
fi

# Function to compare YAML files (ignoring whitespace differences)
compare_yaml() {
    local expected_file="$1"
    local actual_file="$2"
    local description="$3"

    if [ ! -f "$actual_file" ]; then
        echo -e "${RED}FAIL: $description - file not created: $actual_file${NC}"
        return 1
    fi

    # Normalize YAML by sorting and removing empty lines/spaces
    normalize_yaml() {
        python3 -c "
import yaml, sys
try:
    with open('$1', 'r') as f:
        data = yaml.safe_load(f)
    print(yaml.dump(data, default_flow_style=False, sort_keys=True))
except Exception as e:
    print(f'Error processing $1: {e}', file=sys.stderr)
    sys.exit(1)
"
    }

    # Compare normalized YAML
    if normalize_yaml "$expected_file" | diff -u - <(normalize_yaml "$actual_file") > /dev/null; then
        echo -e "${GREEN}PASS: $description${NC}"
        return 0
    else
        echo -e "${RED}FAIL: $description${NC}"
        echo "Expected content (normalized):"
        normalize_yaml "$expected_file"
        echo ""
        echo "Actual content (normalized):"
        normalize_yaml "$actual_file"
        echo ""
        echo "Diff:"
        normalize_yaml "$expected_file" | diff -u - <(normalize_yaml "$actual_file") || true
        return 1
    fi
}

# Function to check file content
check_file_content() {
    local file="$1"
    local pattern="$2"
    local description="$3"

    if [ ! -f "$file" ]; then
        echo -e "${RED}FAIL: $description - file not found: $file${NC}"
        return 1
    fi

    if grep -q "$pattern" "$file"; then
        echo -e "${GREEN}PASS: $description${NC}"
        return 0
    else
        echo -e "${RED}FAIL: $description${NC}"
        echo "Pattern '$pattern' not found in $file"
        echo "File contents:"
        cat "$file"
        return 1
    fi
}

# Function to check file exists
check_file_exists() {
    local file="$1"
    local description="$2"

    if [ -f "$file" ]; then
        echo -e "${GREEN}PASS: $description${NC}"
        return 0
    else
        echo -e "${RED}FAIL: $description - file not found: $file${NC}"
        return 1
    fi
}

#
# Test 1: Basic initialization (without secrets)
#
echo -e "${YELLOW}=== Test 1: Basic initialization (without secrets) ===${NC}"

echo -e "${YELLOW}Cloning test repository...${NC}"
git clone "$TEST_REPO_URL" "$TEST_DIR"
cd "$TEST_DIR"

echo -e "${YELLOW}Running patternizer init...${NC}"
PATTERNIZER_RESOURCES_DIR="$REPO_ROOT" "$PATTERNIZER_BINARY" init

echo -e "${YELLOW}Running verification tests...${NC}"

# Test 1.1: Check values-global.yaml
compare_yaml "$EXPECTED_VALUES_GLOBAL" "values-global.yaml" "values-global.yaml content"

# Test 1.2: Check values-prod.yaml
compare_yaml "$EXPECTED_VALUES_PROD" "values-prod.yaml" "values-prod.yaml content"

# Test 1.3: Check pattern.sh exists and has USE_SECRETS=false
check_file_content "pattern.sh" 'USE_SECRETS:=false' "pattern.sh contains USE_SECRETS=false"

# Test 1.4: Verify pattern.sh is executable
if [ -x "pattern.sh" ]; then
    echo -e "${GREEN}PASS: pattern.sh is executable${NC}"
else
    echo -e "${RED}FAIL: pattern.sh is not executable${NC}"
    exit 1
fi

echo -e "${GREEN}=== Test 1: Basic initialization PASSED ===${NC}"

#
# Test 2: Initialization with secrets
#
echo -e "${YELLOW}=== Test 2: Initialization with secrets ===${NC}"

cd "$REPO_ROOT"  # Go back to repo root
echo -e "${YELLOW}Cloning test repository for secrets test...${NC}"
git clone "$TEST_REPO_URL" "$TEST_DIR_SECRETS"
cd "$TEST_DIR_SECRETS"

echo -e "${YELLOW}Running patternizer init --with-secrets...${NC}"
PATTERNIZER_RESOURCES_DIR="$REPO_ROOT" "$PATTERNIZER_BINARY" init --with-secrets

echo -e "${YELLOW}Running verification tests for secrets...${NC}"

# Test 2.1: Check values-global.yaml (should be same as without secrets)
compare_yaml "$EXPECTED_VALUES_GLOBAL" "values-global.yaml" "values-global.yaml content (with secrets)"

# Test 2.2: Check values-prod.yaml with secrets applications
compare_yaml "$EXPECTED_VALUES_PROD_WITH_SECRETS" "values-prod.yaml" "values-prod.yaml content (with secrets)"

# Test 2.3: Check pattern.sh exists and has USE_SECRETS=true
check_file_content "pattern.sh" 'USE_SECRETS:=true' "pattern.sh contains USE_SECRETS=true"

# Test 2.4: Verify pattern.sh is executable
if [ -x "pattern.sh" ]; then
    echo -e "${GREEN}PASS: pattern.sh is executable (with secrets)${NC}"
else
    echo -e "${RED}FAIL: pattern.sh is not executable (with secrets)${NC}"
    exit 1
fi

# Test 2.5: Check values-secret.yaml.template exists
check_file_exists "values-secret.yaml.template" "values-secret.yaml.template file exists"

echo -e "${GREEN}=== Test 2: Initialization with secrets PASSED ===${NC}"

#
# Test 3: Custom pattern and cluster group names (merging test)
#
echo -e "${YELLOW}=== Test 3: Custom pattern and cluster group names ===${NC}"

cd "$REPO_ROOT"  # Go back to repo root
echo -e "${YELLOW}Cloning test repository for custom names test...${NC}"
git clone "$TEST_REPO_URL" "$TEST_DIR_CUSTOM"
cd "$TEST_DIR_CUSTOM"

echo -e "${YELLOW}Setting up initial values-global.yaml with custom names...${NC}"
cp "$INITIAL_VALUES_GLOBAL_CUSTOM" "values-global.yaml"

echo -e "${YELLOW}Running patternizer init (should preserve custom names)...${NC}"
PATTERNIZER_RESOURCES_DIR="$REPO_ROOT" "$PATTERNIZER_BINARY" init

echo -e "${YELLOW}Running verification tests for custom names...${NC}"

# Test 3.1: Check values-global.yaml preserves custom names and adds multiSourceConfig
compare_yaml "$EXPECTED_VALUES_GLOBAL_CUSTOM" "values-global.yaml" "values-global.yaml content (custom names)"

# Test 3.2: Check custom cluster group file is created with correct content
compare_yaml "$EXPECTED_VALUES_RENAMED_CLUSTER_GROUP" "values-renamed-cluster-group.yaml" "values-renamed-cluster-group.yaml content"

# Test 3.3: Check pattern.sh exists and has USE_SECRETS=false
check_file_content "pattern.sh" 'USE_SECRETS:=false' "pattern.sh contains USE_SECRETS=false (custom names)"

# Test 3.4: Verify pattern.sh is executable
if [ -x "pattern.sh" ]; then
    echo -e "${GREEN}PASS: pattern.sh is executable (custom names)${NC}"
else
    echo -e "${RED}FAIL: pattern.sh is not executable (custom names)${NC}"
    exit 1
fi

echo -e "${GREEN}=== Test 3: Custom pattern and cluster group names PASSED ===${NC}"

echo -e "${GREEN}All integration tests passed!${NC}"

# Clean up
cd "$REPO_ROOT"
rm -rf "$TEST_DIR" "$TEST_DIR_SECRETS" "$TEST_DIR_CUSTOM"
