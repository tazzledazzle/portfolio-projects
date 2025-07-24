#!/bin/bash
set -euo pipefail

echo "🔍 Validating patch syntax..."

validate_patch() {
    local patch_file="$1"
    local patch_name=$(basename "$patch_file" .patch)
    
    echo -n "Checking $patch_name... "
    
    # Extract git diff content and validate
    if sed -n '/^\*\*\* Begin Patch:/,/^\*\*\* End Patch:/p' "$patch_file" | \
       sed '1d;$d' | \
       git apply --check --verbose 2>/dev/null; then
        echo "✅ Valid"
        return 0
    else
        echo "❌ Invalid"
        return 1
    fi
}

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Not in a git repository. Please run 'git init' first."
    exit 1
fi

# Validate all patches
invalid_patches=()
for i in 01 02 03 04 05 06 07 08 09 10; do
    patch_file="pr${i}.patch"
    if [[ -f "$patch_file" ]]; then
        if ! validate_patch "$patch_file"; then
            invalid_patches+=("$patch_file")
        fi
    else
        echo "⚠️  $patch_file not found"
    fi
done

echo ""
if [[ ${#invalid_patches[@]} -eq 0 ]]; then
    echo "🎉 All patches are syntactically valid!"
    echo "Run ./apply_patches.sh to apply them."
else
    echo "❌ Invalid patches found:"
    for patch in "${invalid_patches[@]}"; do
        echo "   - $patch"
    done
    echo ""
    echo "These patches need to be fixed before applying."
fi