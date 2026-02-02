```sh
function ralph2() {
    local logfile="plan/02-e2e-ralph/logs/$(date +%Y%m%d-%H-%M-%S).log"
    echo "ralph2: writing to $logfile"
    env TERM=dumb openhands --headless -f plan/02-e2e-ralph/RALPH.md --json > "$logfile"
}

function save_plan() {
    git add plan && git diff --cached --quiet plan || git commit plan -m "Update plan"
}

```
