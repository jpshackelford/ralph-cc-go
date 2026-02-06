
1. Learn from `plan/03-pop-ralph/logs.*` to determine how we could have made the agent more efficient. What would have helped to know sooner?

2. Do not read the entire logs, they are big. Parse their (mostly) json, make scripts to help you efficiently analyse them. Save these in `plan/04-learn/scripts/` to run with UV.

3. Study failure patterns, record your learning and recommendations in `plan/04-learn/ANALYSIS.md`.

4. When done, commit `plan/04-learn` with your findings.

## UV scripts

Use this pattern to make self contained scripts
```
#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.12"
# dependencies = [
# ]
# ///
```
