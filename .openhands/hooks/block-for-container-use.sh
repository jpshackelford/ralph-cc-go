#!/bin/bash
# PreToolUse hook: Block tool completely because we're using container-use

echo '{"decision": "deny", "reason": "Use container-use for these actions, if you forgot, check AGENTS.md"}'
exit 2  # Exit code 2 = block the operation
