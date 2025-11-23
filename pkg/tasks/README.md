# Task Execution

## Overview

Tasks from base are executed by routing them through existing HTTP handlers. 

## Task Address Format

Tasks use the `TaskAddress` format: `"path:method"`

Examples:
- `"system/status:get"` → GET /system/status
- `"deployments:post"` → POST /deployments
- `"services/abc123:put"` → PUT /services/abc123


## How It Works

1. **Base sends tasks** via `/v1/agent/instances/{instanceId}/status`
2. **Syncer receives tasks** with ID, Type (TaskAddress), and Payload
3. **Executor converts to HTTP request**:
   - Parses `"path:method"` format
   - Creates internal HTTP request with task payload as body
   - Sets `request_id`from task.ID
   - Routes through existing web handlers
4. **Result captured** and queued for next sync

## Logging

Every task execution:
- Uses task ID as `request_id` for all logs
- Generates new `trace_id` for the task execution on handlers
- Logs are structured and auditable in JSON format

Example log:
```json
{
  "level": "info",
  "msg": "executing task",
  "request_id": "task-123",
  "trace_id": "01HXK...",
  "type": "system/status:get"
}
```
Log files are saved to `/var/log/dployrd/dployrd.log`, `C:/ProgramData/dployr/.dployrd/dployrd.log` on Windows

## Authorization

Tasks are routed through existing auth middleware, so:
- Auth requirements are defined once at the route level
- No separate task permission map needed
- Agent role requires `agent` role which is provided in base 
