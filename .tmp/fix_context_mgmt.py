import sys

with open('/home/nexora/internal/agent/coordinator.go', 'r') as f:
    lines = f.readlines()

# Find the line with "if model != nil && model.CatwalkCfg.ContextWindow > 0"
insert_idx = None
for i, line in enumerate(lines):
    if 'if model != nil && model.CatwalkCfg.ContextWindow > 0' in line and i > 560 and i < 570:
        insert_idx = i
        break

if insert_idx is None:
    print("Could not find insertion point")
    sys.exit(1)

# Find the closing brace for the condition (after auto-summarize logs)
end_idx = None
for i in range(insert_idx + 1, insert_idx + 40):
    if 'if strings.HasPrefix(prompt, "/")' in lines[i]:
        end_idx = i - 2  # The line before debug command check
        break

if end_idx is None:
    print("Could not find end point")
    sys.exit(1)

# Prepare replacement text
new_lines = [
    lines[insert_idx] + "\n",
    lines[insert_idx + 1] + "\n",
    lines[insert_idx + 2] + "\n",
    "// Get context management settings (default to Version 3 fixed_80 strategy)\n",
    "ctxMgmt := cfg.Options.ContextManagement\n",
    "if ctxMgmt == nil || ctxMgmt.EnableAuto {\n",
    "\tthreshold := 0.75 // Default: trigger at 75%\n",
    "\tif ctxMgmt != nil {\n",
    "\t\t// Version 1 (ai_cerebras) or Version 2 (ai_same_provider) uses proactive threshold\n",
    "\t\tif ctxMgmt.Strategy == \"ai_cerebras\" || ctxMgmt.Strategy == \"ai_same_provider\" {\n",
    "\t\t\tthreshold = ctxMgmt.ProactiveCompactThreshold\n",
    "\t\t} else {\n",
    "\t\t\t// Version 3 (fixed_80) uses compact threshold\n",
    "\t\t\tthreshold = ctxMgmt.CompactThreshold\n",
    "\t\t}\n",
    "\t}\n",
    "\n",
    "\t\t\t\"threshold\", threshold,\n",
    "\t\t\t\"strategy\", ctxMgmt.Strategy,\n",
]

# Replace lines from condition check to before debug command
old_text = ''.join(lines[insert_idx+2:end_idx])
new_text = ''.join(new_lines)

# Write back with updated content
lines[insert_idx+2:end_idx] = [new_text]

with open('/home/nexora/internal/agent/coordinator.go', 'w') as f:
    f.writelines(lines)

print(f"Updated lines {insert_idx+2} to {end_idx-1}")
