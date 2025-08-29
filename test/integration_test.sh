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
TEST_DIR_UPGRADE="/tmp/patternizer-integration-test-upgrade"
TEST_DIR_UPGRADE_INCLUDE="/tmp/patternizer-integration-test-upgrade-include"
TEST_DIR_UPGRADE_REPLACE="/tmp/patternizer-integration-test-upgrade-replace"
TEST_DIR_UPGRADE_NOMAKEFILE="/tmp/patternizer-integration-test-upgrade-nomakefile"
REPO_NAME=$(basename -s .git "$TEST_REPO_URL")
SHARED_CLONE_PARENT="/tmp/patternizer-shared-clone"
SHARED_REPO_DIR="$SHARED_CLONE_PARENT/$REPO_NAME"

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
if [ -d "$TEST_DIR_UPGRADE" ]; then
    rm -rf "$TEST_DIR_UPGRADE"
fi
if [ -d "$TEST_DIR_UPGRADE_INCLUDE" ]; then
    rm -rf "$TEST_DIR_UPGRADE_INCLUDE"
fi
if [ -d "$TEST_DIR_UPGRADE_REPLACE" ]; then
    rm -rf "$TEST_DIR_UPGRADE_REPLACE"
fi
if [ -d "$TEST_DIR_UPGRADE_NOMAKEFILE" ]; then
    rm -rf "$TEST_DIR_UPGRADE_NOMAKEFILE"
fi

# Clean up any previous shared clone
if [ -d "$SHARED_CLONE_PARENT" ]; then
    rm -rf "$SHARED_CLONE_PARENT"
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
EXPECTED_MAKEFILE_COMMON="$PATTERNIZER_RESOURCES_DIR/Makefile-common"
EXPECTED_PATTERN_SH="$PATTERNIZER_RESOURCES_DIR/pattern.sh"
EXPECTED_VALUES_SECRET_TEMPLATE="$PATTERNIZER_RESOURCES_DIR/values-secret.yaml.template"

# Check if patternizer binary exists and is executable
if [ ! -x "$PATTERNIZER_BINARY" ]; then
    echo -e "${RED}ERROR: Patternizer binary not found or not executable at: $PATTERNIZER_BINARY${NC}"
    exit 1
fi

# Perform a single shallow clone of the source repository into a shared location
mkdir -p "$SHARED_CLONE_PARENT"
cd "$SHARED_CLONE_PARENT"
git clone --depth 1 "$TEST_REPO_URL"
cd "$REPO_ROOT"

# Helper to prepare a test repository copy and cd into it
prepare_and_enter_repo() {
    local dest_dir="$1"
    local header_msg="$2"
    test_header "$header_msg"
    cd "$REPO_ROOT"
    mkdir -p "$dest_dir"
    rm -rf "${dest_dir:?}/${REPO_NAME:?}"
    cp -a "$SHARED_REPO_DIR" "$dest_dir/"
    cd "$dest_dir/$REPO_NAME"
}

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

# Function to print test section headers
test_header() {
    echo -e "${YELLOW}$1${NC}"
}

# Function to print test pass messages
test_pass() {
    echo -e "${GREEN}PASS: $1${NC}"
}

# Function to print test fail messages and exit
test_fail() {
    echo -e "${RED}FAIL: $1${NC}"
    exit 1
}

# Function to compare two files exactly with diff, showing differences on failure
compare_files() {
    local expected_file="$1"
    local actual_file="$2"
    local description="$3"

    if [ ! -f "$actual_file" ]; then
        test_fail "$description - file not created: $actual_file"
    fi

    if [ ! -f "$expected_file" ]; then
        test_fail "$description - expected file not found: $expected_file"
    fi

    if diff "$expected_file" "$actual_file" > /dev/null; then
        test_pass "$description"
        return 0
    else
        echo -e "${RED}FAIL: $description${NC}"
        echo "Expected file: $expected_file"
        echo "Actual file: $actual_file"
        echo "Diff:"
        diff "$expected_file" "$actual_file" || true
        exit 1
    fi
}

# Function to check file exists
check_file_exists() {
    local file="$1"
    local description="$2"

    if [ -f "$file" ]; then
        test_pass "$description"
        return 0
    else
        test_fail "$description - file not found: $file"
    fi
}

#
# Test 1: Basic initialization (without secrets)
#
prepare_and_enter_repo "$TEST_DIR" "=== Test 1: Basic initialization (without secrets) ==="

test_header "Running patternizer init..."
"$PATTERNIZER_BINARY" init

test_header "Running verification tests..."

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
compare_files "$EXPECTED_MAKEFILE" "Makefile" "Makefile has expected content (init without secrets)"

# Test 1.5: Check Makefile-common has exact expected content
compare_files "$EXPECTED_MAKEFILE_COMMON" "Makefile-common" "Makefile-common has expected content (init without secrets)"

test_pass "=== Test 1: Basic initialization PASSED ==="

#
# Test 2: Initialization with secrets
#
prepare_and_enter_repo "$TEST_DIR_SECRETS" "=== Test 2: Initialization with secrets ==="

test_header "Running patternizer init --with-secrets..."
"$PATTERNIZER_BINARY" init --with-secrets

test_header "Running verification tests for secrets..."

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
compare_files "$EXPECTED_VALUES_SECRET_TEMPLATE" "values-secret.yaml.template" "values-secret.yaml.template has expected content"

# Test 2.5: Check Makefile has exact expected content
compare_files "$EXPECTED_MAKEFILE" "Makefile" "Makefile has expected content (init with secrets)"

# Test 2.6: Check Makefile-common has exact expected content
compare_files "$EXPECTED_MAKEFILE_COMMON" "Makefile-common" "Makefile-common has expected content (init with secrets)"

test_pass "=== Test 2: Initialization with secrets PASSED ==="

#
# Test 3: Custom pattern and cluster group names (merging test with secrets)
#
prepare_and_enter_repo "$TEST_DIR_CUSTOM" "=== Test 3: Custom pattern and cluster group names (with secrets) ==="

test_header "Setting up initial values-global.yaml with custom names..."
cp "$INITIAL_VALUES_GLOBAL_CUSTOM" "values-global.yaml"

test_header "Running patternizer init --with-secrets (should preserve custom names)..."
"$PATTERNIZER_BINARY" init --with-secrets

test_header "Running verification tests for custom names..."

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
compare_files "$EXPECTED_VALUES_SECRET_TEMPLATE" "values-secret.yaml.template" "values-secret.yaml.template has expected content (custom names)"

# Test 3.5: Check Makefile has exact expected content
compare_files "$EXPECTED_MAKEFILE" "Makefile" "Makefile has expected content (custom names with secrets)"

# Test 3.6: Check Makefile-common has exact expected content
compare_files "$EXPECTED_MAKEFILE_COMMON" "Makefile-common" "Makefile-common has expected content (custom names with secrets)"

test_pass "=== Test 3: Custom pattern and cluster group names (with secrets) PASSED ==="

#
# Test 4: Sequential execution (init followed by init --with-secrets)
#
prepare_and_enter_repo "$TEST_DIR_SEQUENTIAL" "=== Test 4: Sequential execution (init + init --with-secrets) ==="

test_header "Running patternizer init (first)..."
"$PATTERNIZER_BINARY" init

test_header "Running patternizer init --with-secrets (second)..."
"$PATTERNIZER_BINARY" init --with-secrets

test_header "Running verification tests for sequential execution..."

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
compare_files "$EXPECTED_VALUES_SECRET_TEMPLATE" "values-secret.yaml.template" "values-secret.yaml.template has expected content (sequential)"

# Test 4.5: Check Makefile has exact expected content
compare_files "$EXPECTED_MAKEFILE" "Makefile" "Makefile has expected content (sequential execution)"

# Test 4.6: Check Makefile-common has exact expected content
compare_files "$EXPECTED_MAKEFILE_COMMON" "Makefile-common" "Makefile-common has expected content (sequential execution)"

test_pass "=== Test 4: Sequential execution PASSED ==="

#
# Test 5: File overwrite behavior with existing custom files
#
prepare_and_enter_repo "$TEST_DIR_OVERWRITE" "=== Test 5: File overwrite behavior with existing custom files ==="

test_header "Setting up existing custom files..."

# Copy initial files to set up the test scenario
cp "$INITIAL_VALUES_GLOBAL_OVERWRITE" "values-global.yaml"
cp "$INITIAL_VALUES_CUSTOM_CLUSTER_OVERWRITE" "values-custom-cluster.yaml"
cp "$INITIAL_MAKEFILE_OVERWRITE" "Makefile"
cp "$INITIAL_MAKEFILE_PATTERN_OVERWRITE" "Makefile-common"
cp "$INITIAL_PATTERN_SH_OVERWRITE" "pattern.sh"
cp "$INITIAL_VALUES_SECRET_TEMPLATE_OVERWRITE" "values-secret.yaml.template"

# Make pattern.sh executable to match real scenarios
chmod +x "pattern.sh"

test_header "Running patternizer init --with-secrets..."
"$PATTERNIZER_BINARY" init --with-secrets

test_header "Verifying file overwrite behavior..."

# Test 5.1: values-global.yaml should preserve custom fields and merge with defaults
compare_yaml "$EXPECTED_VALUES_GLOBAL_OVERWRITE" "values-global.yaml" "values-global.yaml content (preserves custom fields with --with-secrets)"

# Test 5.2: values-custom-cluster.yaml should preserve custom fields and merge with defaults
compare_yaml "$EXPECTED_VALUES_CUSTOM_CLUSTER_OVERWRITE" "values-custom-cluster.yaml" "values-custom-cluster.yaml content (preserves custom fields)"

# Test 5.3: Makefile should NOT be overwritten
compare_files "$INITIAL_MAKEFILE_OVERWRITE" "Makefile" "Makefile was not overwritten (content preserved)"

# Test 5.4: Makefile-common SHOULD be overwritten with exact expected content
compare_files "$EXPECTED_MAKEFILE_COMMON" "Makefile-common" "Makefile-common was overwritten with correct content"

# Test 5.5: pattern.sh SHOULD be overwritten with exact expected content and be executable
compare_files "$EXPECTED_PATTERN_SH" "pattern.sh" "pattern.sh was overwritten with correct content"

# Verify it's executable
if [ -x "pattern.sh" ]; then
    test_pass "pattern.sh is executable"
else
    test_fail "pattern.sh is not executable"
fi

# Test 5.6: values-secret.yaml.template should NOT be overwritten
compare_files "$INITIAL_VALUES_SECRET_TEMPLATE_OVERWRITE" "values-secret.yaml.template" "values-secret.yaml.template was not overwritten (content preserved)"

test_pass "=== Test 5: File overwrite behavior PASSED ==="

#
# Test 6: Mixed file overwrite behavior (some files exist, some don't)
#
prepare_and_enter_repo "$TEST_DIR_MIXED" "=== Test 6: Mixed file overwrite behavior ==="

test_header "Setting up partial existing files..."

# Only create some files to test mixed scenarios

# Copy initial files for mixed scenario
cp "$INITIAL_MAKEFILE_OVERWRITE" "Makefile"
cp "$INITIAL_VALUES_SECRET_TEMPLATE_OVERWRITE" "values-secret.yaml.template"

# Don't create values-global.yaml, values-prod.yaml (should be created)
# Don't create Makefile-common, pattern.sh (should be created/overwritten)

test_header "Running patternizer init --with-secrets on mixed repository..."
"$PATTERNIZER_BINARY" init --with-secrets

test_header "Verifying mixed overwrite behavior..."

# Test 6.1: Files that should be created with exact expected content
check_file_exists "values-global.yaml" "values-global.yaml created when missing"
check_file_exists "values-prod.yaml" "values-prod.yaml created when missing"

compare_files "$EXPECTED_MAKEFILE_COMMON" "Makefile-common" "Makefile-common created with correct content"

compare_files "$EXPECTED_PATTERN_SH" "pattern.sh" "pattern.sh created with correct content"

# Test 6.2: Files that should be preserved
compare_files "$INITIAL_MAKEFILE_OVERWRITE" "Makefile" "Existing Makefile preserved in mixed scenario"

compare_files "$INITIAL_VALUES_SECRET_TEMPLATE_OVERWRITE" "values-secret.yaml.template" "Existing values-secret.yaml.template preserved in mixed scenario"

# Test 6.3: Verify pattern.sh is executable
if [ -x "pattern.sh" ]; then
    test_pass "pattern.sh is executable in mixed scenario"
else
    test_fail "pattern.sh is not executable in mixed scenario"
fi

test_pass "=== Test 6: Mixed file overwrite behavior PASSED ==="

#
# Test 7: Upgrade without --replace-makefile, inject include on first line
#
prepare_and_enter_repo "$TEST_DIR_UPGRADE" "=== Test 7: Upgrade (no replace, inject include) ==="

# Simulate legacy structure
mkdir -p common
ln -s common/pattern.sh pattern.sh

# Create a simple Makefile without include
cat > Makefile <<'EOF'
all:
	@echo hello
EOF

# Run upgrade
"$PATTERNIZER_BINARY" upgrade

# Verify common/ removed and pattern.sh replaced (not symlink)
if [ -d common ]; then
    test_fail "common directory was not removed by upgrade"
fi
if [ -L pattern.sh ]; then
    test_fail "pattern.sh symlink was not removed by upgrade"
fi

# Verify pattern.sh and Makefile-common contents
compare_files "$EXPECTED_PATTERN_SH" "pattern.sh" "pattern.sh copied during upgrade"
compare_files "$EXPECTED_MAKEFILE_COMMON" "Makefile-common" "Makefile-common copied during upgrade"

# Verify Makefile first line and content
EXPECTED_UPGRADE_MF=$(mktemp)
printf "include Makefile-common\nall:\n\t@echo hello\n" > "$EXPECTED_UPGRADE_MF"
compare_files "$EXPECTED_UPGRADE_MF" "Makefile" "Makefile injected include at first line"
rm -f "$EXPECTED_UPGRADE_MF"

test_pass "=== Test 7: Upgrade (no replace, inject include) PASSED ==="

#
# Test 8: Upgrade without --replace-makefile, include already exists elsewhere
#
prepare_and_enter_repo "$TEST_DIR_UPGRADE_INCLUDE" "=== Test 8: Upgrade (no replace, include present) ==="

# Legacy bits
mkdir -p common
ln -s common/pattern.sh pattern.sh

# Makefile already contains include, not on first line
cat > Makefile <<'EOF'
foo:
	@echo foo
include Makefile-common
bar:
	@echo bar
EOF
cp Makefile /tmp/expected_makefile_include_present

# Run upgrade
"$PATTERNIZER_BINARY" upgrade

# Verify removals and copies
if [ -d common ]; then
    test_fail "common directory was not removed by upgrade (include-present case)"
fi
if [ -L pattern.sh ]; then
    test_fail "pattern.sh symlink was not removed by upgrade (include-present case)"
fi
compare_files "$EXPECTED_PATTERN_SH" "pattern.sh" "pattern.sh copied during upgrade (include-present)"
compare_files "$EXPECTED_MAKEFILE_COMMON" "Makefile-common" "Makefile-common copied during upgrade (include-present)"

# Verify Makefile unchanged
compare_files "/tmp/expected_makefile_include_present" "Makefile" "Makefile unchanged when include already present"
rm -f /tmp/expected_makefile_include_present

test_pass "=== Test 8: Upgrade (no replace, include present) PASSED ==="

#
# Test 9: Upgrade with --replace-makefile replaces Makefile exactly
#
prepare_and_enter_repo "$TEST_DIR_UPGRADE_REPLACE" "=== Test 9: Upgrade (--replace-makefile) ==="

# Create a custom Makefile to be overwritten
echo "custom: ; @echo custom" > Makefile

# Run upgrade with flag
"$PATTERNIZER_BINARY" upgrade --replace-makefile

# Verify Makefile replaced and other files copied
compare_files "$EXPECTED_MAKEFILE" "Makefile" "Makefile replaced during upgrade --replace-makefile"
compare_files "$EXPECTED_MAKEFILE_COMMON" "Makefile-common" "Makefile-common copied during upgrade --replace-makefile"
compare_files "$EXPECTED_PATTERN_SH" "pattern.sh" "pattern.sh copied during upgrade --replace-makefile"

test_pass "=== Test 9: Upgrade (--replace-makefile) PASSED ==="

#
# Test 10: Upgrade without existing Makefile creates default Makefile
#
prepare_and_enter_repo "$TEST_DIR_UPGRADE_NOMAKEFILE" "=== Test 10: Upgrade (no Makefile present) ==="

# Ensure no Makefile exists
rm -f Makefile

# Run upgrade
"$PATTERNIZER_BINARY" upgrade

# Verify Makefile created and matches default
compare_files "$EXPECTED_MAKEFILE" "Makefile" "Makefile created during upgrade when missing"
compare_files "$EXPECTED_MAKEFILE_COMMON" "Makefile-common" "Makefile-common copied during upgrade when missing"
compare_files "$EXPECTED_PATTERN_SH" "pattern.sh" "pattern.sh copied during upgrade when missing"

test_pass "=== Test 10: Upgrade (no Makefile present) PASSED ==="

test_pass "All integration tests passed!"

# Clean up
cd "$REPO_ROOT"
rm -rf "$TEST_DIR" "$TEST_DIR_SECRETS" "$TEST_DIR_CUSTOM" "$TEST_DIR_SEQUENTIAL" "$TEST_DIR_OVERWRITE" "$TEST_DIR_MIXED" "$TEST_DIR_UPGRADE" "$TEST_DIR_UPGRADE_INCLUDE" "$TEST_DIR_UPGRADE_REPLACE" "$TEST_DIR_UPGRADE_NOMAKEFILE" "$SHARED_CLONE_PARENT"
