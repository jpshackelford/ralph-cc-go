# Research only

```sh
function ralph5() {
    local logfile="plan/05-fix-research-ralph/logs/$(date +%Y%m%d-%H-%M-%S).log"
    echo "ralph5: writing to $logfile"
    env TERM=dumb openhands --headless -f plan/05-fix-research-ralph/RALPH.md --json > "$logfile"
}

function save_plan() {
    git add plan && git diff --cached --quiet plan || git commit plan -m "Update plan"
}

```
