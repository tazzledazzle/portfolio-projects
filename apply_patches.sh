#!/bin/bash
set -euo pipefail

echo "🔧 Applying portfolio project patches..."

# Function to apply a single patch with custom format
apply_patch() {
    local patch_file="$1"
    local patch_name=$(basename "$patch_file" .patch)
    
    echo "📦 Applying $patch_name..."
    
    # Extract just the git diff content, skipping the custom markers
    if sed -n '/^\*\*\* Begin Patch:/,/^\*\*\* End Patch:/p' "$patch_file" | \
       sed '1d;$d' | \
       git apply --verbose 2>&1; then
        echo "✅ $patch_name applied successfully"
        return 0
    else
        echo "❌ Failed to apply $patch_name"
        echo "   You may need to resolve conflicts manually"
        return 1
    fi
}

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Not in a git repository. Please run 'git init' first."
    exit 1
fi

# Apply patches in order
failed_patches=()
for i in 01 02 03 04 05 06 07 08 09 10; do
    patch_file="pr${i}.patch"
    if [[ -f "$patch_file" ]]; then
        if ! apply_patch "$patch_file"; then
            failed_patches+=("$patch_file")
        fi
    else
        echo "⚠️  $patch_file not found, skipping..."
    fi
done

# Summary
echo ""
echo "📋 Summary:"
if [[ ${#failed_patches[@]} -eq 0 ]]; then
    echo "🎉 All patches applied successfully!"
    echo ""
    echo "Next steps:"
    echo "1. Review the changes: git status"
    echo "2. Commit the changes: git add . && git commit -m 'Apply portfolio patches'"
    echo "3. Check the project structure and run any setup scripts"
else
    echo "⚠️  Some patches failed to apply:"
    for patch in "${failed_patches[@]}"; do
        echo "   - $patch"
    done
    echo ""
    echo "You may need to:"
    echo "1. Check for file conflicts"
    echo "2. Apply failed patches manually"
    echo "3. Review git status for any issues"
fi