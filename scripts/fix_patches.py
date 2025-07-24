#!/usr/bin/env python3
"""
Fix patch files by ensuring proper Git patch format.
Adds blank lines between different file patches.
"""

import re
import sys
from pathlib import Path

def fix_patch_file(patch_path):
    """Fix a single patch file by adding proper separators."""
    print(f"Fixing {patch_path}...")
    
    with open(patch_path, 'r') as f:
        content = f.read()
    
    # Extract content between markers
    start_marker = "*** Begin Patch:"
    end_marker = "*** End Patch:"
    
    start_idx = content.find(start_marker)
    end_idx = content.find(end_marker)
    
    if start_idx == -1 or end_idx == -1:
        print(f"  ❌ No markers found in {patch_path}")
        return False
    
    # Get the patch content (skip the marker lines)
    patch_content = content[start_idx:end_idx].split('\n')[1:]
    
    # Process lines to add proper separators
    fixed_lines = []
    prev_was_diff = False
    
    for line in patch_content:
        # If this is a new diff --git line and we had a previous diff
        if line.startswith('diff --git') and prev_was_diff:
            # Add a blank line before the new diff
            fixed_lines.append('')
        
        fixed_lines.append(line)
        prev_was_diff = line.startswith('diff --git')
    
    # Remove any trailing empty lines and ensure we end properly
    while fixed_lines and fixed_lines[-1] == '':
        fixed_lines.pop()
    
    # Write the fixed patch
    fixed_content = '\n'.join(fixed_lines)
    
    # Write to a new file
    fixed_path = patch_path.with_suffix('.fixed.patch')
    with open(fixed_path, 'w') as f:
        f.write(fixed_content)
    
    print(f"  ✅ Fixed patch saved as {fixed_path}")
    return True

def main():
    """Fix all patch files."""
    patch_files = list(Path('.').glob('pr*.patch'))
    patch_files.sort()
    
    if not patch_files:
        print("No patch files found!")
        return 1
    
    print(f"Found {len(patch_files)} patch files to fix...")
    
    success_count = 0
    for patch_file in patch_files:
        if fix_patch_file(patch_file):
            success_count += 1
    
    print(f"\n✅ Fixed {success_count}/{len(patch_files)} patch files")
    print("Fixed patches have .fixed.patch extension")
    
    return 0 if success_count == len(patch_files) else 1

if __name__ == '__main__':
    sys.exit(main())