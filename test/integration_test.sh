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
TEST_DIR_SEQUENTIAL="/tmp/patternizer-integration-test-sequential"
TEST_DIR_OVERWRITE="/tmp/patternizer-integration-test-overwrite"
TEST_DIR_MIXED="/tmp/patternizer-integration-test-mixed"

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
if [ -d "$TEST_DIR_SEQUENTIAL" ]; then
    rm -rf "$TEST_DIR_SEQUENTIAL"
fi
if [ -d "$TEST_DIR_OVERWRITE" ]; then
    rm -rf "$TEST_DIR_OVERWRITE"
fi
if [ -d "$TEST_DIR_MIXED" ]; then
    rm -rf "$TEST_DIR_MIXED"
fi

# Convert PATTERNIZER_BINARY to absolute path before changing directories
PATTERNIZER_BINARY=$(realpath "$PATTERNIZER_BINARY")

# Get the absolute path to the repository root (where resource files are located)
REPO_ROOT=$(pwd)

# Export resources directory so patternizer can find resource files
export PATTERNIZER_RESOURCES_DIR="$REPO_ROOT/resources"

# Set absolute paths to expected files
EXPECTED_VALUES_CUSTOM_CLUSTER_OVERWRITE="$REPO_ROOT/test/expected_values_custom_cluster_overwrite.yaml"
EXPECTED_VALUES_GLOBAL_CUSTOM="$REPO_ROOT/test/expected_values_global_custom.yaml"
EXPECTED_VALUES_GLOBAL_OVERWRITE="$REPO_ROOT/test/expected_values_global_overwrite.yaml"
EXPECTED_VALUES_GLOBAL_WITH_SECRETS="$REPO_ROOT/test/expected_values_global_with_secrets.yaml"
EXPECTED_VALUES_GLOBAL="$REPO_ROOT/test/expected_values_global.yaml"
EXPECTED_VALUES_PROD_WITH_SECRETS="$REPO_ROOT/test/expected_values_prod_with_secrets.yaml"
EXPECTED_VALUES_PROD="$REPO_ROOT/test/expected_values_prod.yaml"
EXPECTED_VALUES_RENAMED_CLUSTER_GROUP="$REPO_ROOT/test/expected_values_renamed_cluster_group.yaml"
INITIAL_MAKEFILE_OVERWRITE="$REPO_ROOT/test/initial_makefile_overwrite"
INITIAL_MAKEFILE_PATTERN_OVERWRITE="$REPO_ROOT/test/initial_makefile_pattern_overwrite"
INITIAL_PATTERN_SH_OVERWRITE="$REPO_ROOT/test/initial_pattern_sh_overwrite"
INITIAL_VALUES_CUSTOM_CLUSTER_OVERWRITE="$REPO_ROOT/test/initial_values_custom_cluster_overwrite.yaml"
INITIAL_VALUES_GLOBAL_CUSTOM="$REPO_ROOT/test/initial_values_global_custom.yaml"
INITIAL_VALUES_GLOBAL_OVERWRITE="$REPO_ROOT/test/initial_values_global_overwrite.yaml"
INITIAL_VALUES_SECRET_TEMPLATE_OVERWRITE="$REPO_ROOT/test/initial_values_secret_template_overwrite.yaml"

# Set paths for expected resource files
EXPECTED_MAKEFILE="$PATTERNIZER_RESOURCES_DIR/Makefile"
EXPECTED_MAKEFILE_PATTERN="$PATTERNIZER_RESOURCES_DIR/Makefile-pattern"
EXPECTED_PATTERN_SH="$PATTERNIZER_RESOURCES_DIR/pattern.sh"
EXPECTED_VALUES_SECRET_TEMPLATE="$PATTERNIZER_RESOURCES_DIR/values-secret.yaml.template"

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
"$PATTERNIZER_BINARY" init

echo -e "${YELLOW}Running verification tests...${NC}"

# Test 1.1: Check values-global.yaml
compare_yaml "$EXPECTED_VALUES_GLOBAL" "values-global.yaml" "values-global.yaml content"

# Test 1.2: Check values-prod.yaml
compare_yaml "$EXPECTED_VALUES_PROD" "values-prod.yaml" "values-prod.yaml content"

# Test 1.3: Verify pattern.sh is executable
if [ -x "pattern.sh" ]; then
    echo -e "${GREEN}PASS: pattern.sh is executable${NC}"
else
    echo -e "${RED}FAIL: pattern.sh is not executable${NC}"
    exit 1
fi

# Test 1.4: Check Makefile has exact expected content
if diff "$EXPECTED_MAKEFILE" "Makefile" > /dev/null; then
    echo -e "${GREEN}PASS: Makefile has expected content (init without secrets)${NC}"
else
    echo -e "${RED}FAIL: Makefile content doesn't match expected (init without secrets)${NC}"
    exit 1
fi

# Test 1.5: Check Makefile-pattern has exact expected content
if diff "$EXPECTED_MAKEFILE_PATTERN" "Makefile-pattern" > /dev/null; then
    echo -e "${GREEN}PASS: Makefile-pattern has expected content (init without secrets)${NC}"
else
    echo -e "${RED}FAIL: Makefile-pattern content doesn't match expected (init without secrets)${NC}"
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
"$PATTERNIZER_BINARY" init --with-secrets

echo -e "${YELLOW}Running verification tests for secrets...${NC}"

# Test 2.1: Check values-global.yaml (secretLoader.disabled should be false with secrets)
compare_yaml "$EXPECTED_VALUES_GLOBAL_WITH_SECRETS" "values-global.yaml" "values-global.yaml content (with secrets)"

# Test 2.2: Check values-prod.yaml with secrets applications
compare_yaml "$EXPECTED_VALUES_PROD_WITH_SECRETS" "values-prod.yaml" "values-prod.yaml content (with secrets)"

# Test 2.3: Verify pattern.sh is executable
if [ -x "pattern.sh" ]; then
    echo -e "${GREEN}PASS: pattern.sh is executable (with secrets)${NC}"
else
    echo -e "${RED}FAIL: pattern.sh is not executable (with secrets)${NC}"
    exit 1
fi

# Test 2.4: Check values-secret.yaml.template has exact expected content
if diff "$EXPECTED_VALUES_SECRET_TEMPLATE" "values-secret.yaml.template" > /dev/null; then
    echo -e "${GREEN}PASS: values-secret.yaml.template has expected content${NC}"
else
    echo -e "${RED}FAIL: values-secret.yaml.template content doesn't match expected${NC}"
    exit 1
fi

# Test 2.5: Check Makefile has exact expected content
if diff "$EXPECTED_MAKEFILE" "Makefile" > /dev/null; then
    echo -e "${GREEN}PASS: Makefile has expected content (init with secrets)${NC}"
else
    echo -e "${RED}FAIL: Makefile content doesn't match expected (init with secrets)${NC}"
    exit 1
fi

# Test 2.6: Check Makefile-pattern has exact expected content
if diff "$EXPECTED_MAKEFILE_PATTERN" "Makefile-pattern" > /dev/null; then
    echo -e "${GREEN}PASS: Makefile-pattern has expected content (init with secrets)${NC}"
else
    echo -e "${RED}FAIL: Makefile-pattern content doesn't match expected (init with secrets)${NC}"
    exit 1
fi

echo -e "${GREEN}=== Test 2: Initialization with secrets PASSED ===${NC}"

#
# Test 3: Custom pattern and cluster group names (merging test with secrets)
#
echo -e "${YELLOW}=== Test 3: Custom pattern and cluster group names (with secrets) ===${NC}"

cd "$REPO_ROOT"  # Go back to repo root
echo -e "${YELLOW}Cloning test repository for custom names test...${NC}"
git clone "$TEST_REPO_URL" "$TEST_DIR_CUSTOM"
cd "$TEST_DIR_CUSTOM"

echo -e "${YELLOW}Setting up initial values-global.yaml with custom names...${NC}"
cp "$INITIAL_VALUES_GLOBAL_CUSTOM" "values-global.yaml"

echo -e "${YELLOW}Running patternizer init --with-secrets (should preserve custom names)...${NC}"
"$PATTERNIZER_BINARY" init --with-secrets

echo -e "${YELLOW}Running verification tests for custom names...${NC}"

# Test 3.1: Check values-global.yaml preserves custom names and adds multiSourceConfig
compare_yaml "$EXPECTED_VALUES_GLOBAL_CUSTOM" "values-global.yaml" "values-global.yaml content (custom names)"

# Test 3.2: Check custom cluster group file is created with correct content
compare_yaml "$EXPECTED_VALUES_RENAMED_CLUSTER_GROUP" "values-renamed-cluster-group.yaml" "values-renamed-cluster-group.yaml content"

# Test 3.3: Verify pattern.sh is executable
if [ -x "pattern.sh" ]; then
    echo -e "${GREEN}PASS: pattern.sh is executable (custom names)${NC}"
else
    echo -e "${RED}FAIL: pattern.sh is not executable (custom names)${NC}"
    exit 1
fi

# Test 3.4: Check values-secret.yaml.template has exact expected content
if diff "$EXPECTED_VALUES_SECRET_TEMPLATE" "values-secret.yaml.template" > /dev/null; then
    echo -e "${GREEN}PASS: values-secret.yaml.template has expected content (custom names)${NC}"
else
    echo -e "${RED}FAIL: values-secret.yaml.template content doesn't match expected (custom names)${NC}"
    exit 1
fi

# Test 3.5: Check Makefile has exact expected content
if diff "$EXPECTED_MAKEFILE" "Makefile" > /dev/null; then
    echo -e "${GREEN}PASS: Makefile has expected content (custom names with secrets)${NC}"
else
    echo -e "${RED}FAIL: Makefile content doesn't match expected (custom names with secrets)${NC}"
    exit 1
fi

# Test 3.6: Check Makefile-pattern has exact expected content
if diff "$EXPECTED_MAKEFILE_PATTERN" "Makefile-pattern" > /dev/null; then
    echo -e "${GREEN}PASS: Makefile-pattern has expected content (custom names with secrets)${NC}"
else
    echo -e "${RED}FAIL: Makefile-pattern content doesn't match expected (custom names with secrets)${NC}"
    exit 1
fi

echo -e "${GREEN}=== Test 3: Custom pattern and cluster group names (with secrets) PASSED ===${NC}"

#
# Test 4: Sequential execution (init followed by init --with-secrets)
#
echo -e "${YELLOW}=== Test 4: Sequential execution (init + init --with-secrets) ===${NC}"

cd "$REPO_ROOT"  # Go back to repo root

echo -e "${YELLOW}Cloning test repository for sequential test...${NC}"
git clone "$TEST_REPO_URL" "$TEST_DIR_SEQUENTIAL"
cd "$TEST_DIR_SEQUENTIAL"

echo -e "${YELLOW}Running patternizer init (first)...${NC}"
"$PATTERNIZER_BINARY" init

echo -e "${YELLOW}Running patternizer init --with-secrets (second)...${NC}"
"$PATTERNIZER_BINARY" init --with-secrets

echo -e "${YELLOW}Running verification tests for sequential execution...${NC}"

# Test 4.1: Check values-global.yaml (should have secretLoader.disabled=false after --with-secrets)
compare_yaml "$EXPECTED_VALUES_GLOBAL_WITH_SECRETS" "values-global.yaml" "values-global.yaml content (sequential)"

# Test 4.2: Check values-prod.yaml matches the --with-secrets output
compare_yaml "$EXPECTED_VALUES_PROD_WITH_SECRETS" "values-prod.yaml" "values-prod.yaml content (sequential, should match --with-secrets)"

# Test 4.3: Verify pattern.sh is executable
if [ -x "pattern.sh" ]; then
    echo -e "${GREEN}PASS: pattern.sh is executable (sequential)${NC}"
else
    echo -e "${RED}FAIL: pattern.sh is not executable (sequential)${NC}"
    exit 1
fi

# Test 4.4: Check values-secret.yaml.template has exact expected content
if diff "$EXPECTED_VALUES_SECRET_TEMPLATE" "values-secret.yaml.template" > /dev/null; then
    echo -e "${GREEN}PASS: values-secret.yaml.template has expected content (sequential)${NC}"
else
    echo -e "${RED}FAIL: values-secret.yaml.template content doesn't match expected (sequential)${NC}"
    exit 1
fi

# Test 4.5: Check Makefile has exact expected content
if diff "$EXPECTED_MAKEFILE" "Makefile" > /dev/null; then
    echo -e "${GREEN}PASS: Makefile has expected content (sequential execution)${NC}"
else
    echo -e "${RED}FAIL: Makefile content doesn't match expected (sequential execution)${NC}"
    exit 1
fi

# Test 4.6: Check Makefile-pattern has exact expected content
if diff "$EXPECTED_MAKEFILE_PATTERN" "Makefile-pattern" > /dev/null; then
    echo -e "${GREEN}PASS: Makefile-pattern has expected content (sequential execution)${NC}"
else
    echo -e "${RED}FAIL: Makefile-pattern content doesn't match expected (sequential execution)${NC}"
    exit 1
fi

echo -e "${GREEN}=== Test 4: Sequential execution PASSED ===${NC}"

#
# Test 5: File overwrite behavior with existing custom files
#
echo -e "${YELLOW}=== Test 5: File overwrite behavior with existing custom files ===${NC}"

cd "$REPO_ROOT"  # Go back to repo root

echo -e "${YELLOW}Cloning test repository for overwrite behavior test...${NC}"
git clone "$TEST_REPO_URL" "$TEST_DIR_OVERWRITE"
cd "$TEST_DIR_OVERWRITE"

echo -e "${YELLOW}Setting up existing custom files...${NC}"

# Copy initial files to set up the test scenario
cp "$INITIAL_VALUES_GLOBAL_OVERWRITE" "values-global.yaml"
cp "$INITIAL_VALUES_CUSTOM_CLUSTER_OVERWRITE" "values-custom-cluster.yaml"
cp "$INITIAL_MAKEFILE_OVERWRITE" "Makefile"
cp "$INITIAL_MAKEFILE_PATTERN_OVERWRITE" "Makefile-pattern"
cp "$INITIAL_PATTERN_SH_OVERWRITE" "pattern.sh"
cp "$INITIAL_VALUES_SECRET_TEMPLATE_OVERWRITE" "values-secret.yaml.template"

# Make pattern.sh executable to match real scenarios
chmod +x "pattern.sh"

echo -e "${YELLOW}Running patternizer init --with-secrets...${NC}"
"$PATTERNIZER_BINARY" init --with-secrets

echo -e "${YELLOW}Verifying file overwrite behavior...${NC}"

# Test 5.1: values-global.yaml should preserve custom fields and merge with defaults
compare_yaml "$EXPECTED_VALUES_GLOBAL_OVERWRITE" "values-global.yaml" "values-global.yaml content (preserves custom fields with --with-secrets)"

# Test 5.2: values-custom-cluster.yaml should preserve custom fields and merge with defaults
compare_yaml "$EXPECTED_VALUES_CUSTOM_CLUSTER_OVERWRITE" "values-custom-cluster.yaml" "values-custom-cluster.yaml content (preserves custom fields)"

# Test 5.3: Makefile should NOT be overwritten
if diff "$INITIAL_MAKEFILE_OVERWRITE" "Makefile" > /dev/null; then
    echo -e "${GREEN}PASS: Makefile was not overwritten (content preserved)${NC}"
else
    echo -e "${RED}FAIL: Makefile was overwritten but should have been preserved${NC}"
    echo "Expected (initial):"
    cat "$INITIAL_MAKEFILE_OVERWRITE"
    echo ""
    echo "Actual:"
    cat "Makefile"
    echo ""
    exit 1
fi

# Test 5.4: Makefile-pattern SHOULD be overwritten with exact expected content
if diff "$EXPECTED_MAKEFILE_PATTERN" "Makefile-pattern" > /dev/null; then
    echo -e "${GREEN}PASS: Makefile-pattern was overwritten with correct content${NC}"
else
    echo -e "${RED}FAIL: Makefile-pattern doesn't have expected content after overwrite${NC}"
    exit 1
fi

# Test 5.5: pattern.sh SHOULD be overwritten with exact expected content and be executable
if diff "$EXPECTED_PATTERN_SH" "pattern.sh" > /dev/null; then
    echo -e "${GREEN}PASS: pattern.sh was overwritten with correct content${NC}"
else
    echo -e "${RED}FAIL: pattern.sh doesn't have expected content after overwrite${NC}"
    exit 1
fi

# Verify it's executable
if [ -x "pattern.sh" ]; then
    echo -e "${GREEN}PASS: pattern.sh is executable${NC}"
else
    echo -e "${RED}FAIL: pattern.sh is not executable${NC}"
    exit 1
fi

# Test 5.6: values-secret.yaml.template should NOT be overwritten
if diff "$INITIAL_VALUES_SECRET_TEMPLATE_OVERWRITE" "values-secret.yaml.template" > /dev/null; then
    echo -e "${GREEN}PASS: values-secret.yaml.template was not overwritten (content preserved)${NC}"
else
    echo -e "${RED}FAIL: values-secret.yaml.template was overwritten but should have been preserved${NC}"
    echo "Expected (initial):"
    cat "$INITIAL_VALUES_SECRET_TEMPLATE_OVERWRITE"
    echo ""
    echo "Actual:"
    cat "values-secret.yaml.template"
    echo ""
    exit 1
fi

echo -e "${GREEN}=== Test 5: File overwrite behavior PASSED ===${NC}"

#
# Test 6: Mixed file overwrite behavior (some files exist, some don't)
#
echo -e "${YELLOW}=== Test 6: Mixed file overwrite behavior ===${NC}"

cd "$REPO_ROOT"  # Go back to repo root

echo -e "${YELLOW}Cloning test repository for mixed scenario...${NC}"
git clone "$TEST_REPO_URL" "$TEST_DIR_MIXED"
cd "$TEST_DIR_MIXED"

echo -e "${YELLOW}Setting up partial existing files...${NC}"

# Only create some files to test mixed scenarios

# Copy initial files for mixed scenario
cp "$INITIAL_MAKEFILE_OVERWRITE" "Makefile"
cp "$INITIAL_VALUES_SECRET_TEMPLATE_OVERWRITE" "values-secret.yaml.template"

# Don't create values-global.yaml, values-prod.yaml (should be created)
# Don't create Makefile-pattern, pattern.sh (should be created/overwritten)

echo -e "${YELLOW}Running patternizer init --with-secrets on mixed repository...${NC}"
"$PATTERNIZER_BINARY" init --with-secrets

echo -e "${YELLOW}Verifying mixed overwrite behavior...${NC}"

# Test 6.1: Files that should be created with exact expected content
check_file_exists "values-global.yaml" "values-global.yaml created when missing"
check_file_exists "values-prod.yaml" "values-prod.yaml created when missing"

if diff "$EXPECTED_MAKEFILE_PATTERN" "Makefile-pattern" > /dev/null; then
    echo -e "${GREEN}PASS: Makefile-pattern created with correct content${NC}"
else
    echo -e "${RED}FAIL: Makefile-pattern doesn't have expected content${NC}"
    exit 1
fi

if diff "$EXPECTED_PATTERN_SH" "pattern.sh" > /dev/null; then
    echo -e "${GREEN}PASS: pattern.sh created with correct content${NC}"
else
    echo -e "${RED}FAIL: pattern.sh doesn't have expected content${NC}"
    exit 1
fi

# Test 6.2: Files that should be preserved
if diff "$INITIAL_MAKEFILE_OVERWRITE" "Makefile" > /dev/null; then
    echo -e "${GREEN}PASS: Existing Makefile preserved in mixed scenario${NC}"
else
    echo -e "${RED}FAIL: Existing Makefile was changed in mixed scenario${NC}"
    exit 1
fi

if diff "$INITIAL_VALUES_SECRET_TEMPLATE_OVERWRITE" "values-secret.yaml.template" > /dev/null; then
    echo -e "${GREEN}PASS: Existing values-secret.yaml.template preserved in mixed scenario${NC}"
else
    echo -e "${RED}FAIL: Existing values-secret.yaml.template was changed in mixed scenario${NC}"
    exit 1
fi

# Test 6.3: Verify pattern.sh is executable
if [ -x "pattern.sh" ]; then
    echo -e "${GREEN}PASS: pattern.sh is executable in mixed scenario${NC}"
else
    echo -e "${RED}FAIL: pattern.sh is not executable in mixed scenario${NC}"
    exit 1
fi

echo -e "${GREEN}=== Test 6: Mixed file overwrite behavior PASSED ===${NC}"

echo -e "${GREEN}All integration tests passed!${NC}"

# Clean up
cd "$REPO_ROOT"
rm -rf "$TEST_DIR" "$TEST_DIR_SECRETS" "$TEST_DIR_CUSTOM" "$TEST_DIR_SEQUENTIAL" "$TEST_DIR_OVERWRITE" "$TEST_DIR_MIXED"
