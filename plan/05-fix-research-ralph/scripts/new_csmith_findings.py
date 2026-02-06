#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.12"
# dependencies = []
# ///
"""Find csmith report IDs not yet mentioned in PLAN.md."""

import re
from pathlib import Path

def main():
    base = Path(__file__).parent.parent.parent.parent
    reports_dir = base / "csmith-reports"
    plan_file = Path(__file__).parent.parent / "PLAN.md"
    
    # Get all report IDs from filenames
    report_ids = set()
    for f in reports_dir.glob("report-*.md"):
        # Extract timestamp ID like 20260205-225448
        match = re.search(r'report-(\d{8}-\d{6})\.md$', f.name)
        if match:
            report_ids.add(match.group(1))
    
    # Get IDs already mentioned in PLAN.md
    mentioned_ids = set()
    if plan_file.exists():
        content = plan_file.read_text()
        mentioned_ids = set(re.findall(r'\d{8}-\d{6}', content))
    
    # Find new ones
    new_ids = sorted(report_ids - mentioned_ids)
    
    if new_ids:
        print("New findings not in PLAN.md:")
        for id in new_ids:
            print(f"  - [ ] {id}")
    else:
        print("No new findings.")

if __name__ == "__main__":
    main()
