Executes bash commands with automatic background conversion for long-running tasks.

<cross_platform>
Uses mvdan/sh interpreter (Bash-compatible on all platforms including Windows).
Use forward slashes for paths: "ls C:/foo/bar" not "ls C:\foo\bar".
</cross_platform>

<execution_steps>
1. Directory Verification: Use LS tool to verify parent exists before creating files
2. Security Check: No restrictions - explain to user. Safe read-only commands execute without prompts
3. Command Execution: Execute with proper quoting, capture output
4. Auto-Background: Commands >1 minute automatically move to background, return shell ID
5. Output Processing: Truncate if exceeds {{ .MaxOutputLength }} characters
6. Return Result: Include errors, metadata with <cwd></cwd> tags
</execution_steps>

<usage_notes>
- Command required, working_dir optional (defaults to current directory)
- IMPORTANT: Use Grep/Glob/Agent tools instead of 'find'/'grep'. Use View/LS instead of 'cat'/'head'/'tail'/'ls'
- Chain with ';' or '&&', avoid newlines except in quoted strings
- Each command runs in independent shell (no state persistence)
- Prefer absolute paths over 'cd'
</usage_notes>

<background_execution>
- Set run_in_background=true for separate background shell
- Returns shell ID for management
- Use job_output tool to view output, job_kill to terminate
- NEVER use `&` at end of commands - use run_in_background parameter
- Background: Long servers (npm start, python -m http.server), watch tasks, continuous processes
- Foreground: Builds, tests, git ops, file ops, short scripts
</background_execution>

<git_commits>
When user asks to create git commit:

1. Run: git status, git diff, git log (single message, three tool_use blocks)
2. Add relevant untracked files to staging
3. Analyze in <commit_analysis> tags:
   - List changed files, summarize nature (feature/bug fix/refactor/docs)
   - Draft concise (1-2 sentences) focusing on "why" not "what"
   - Use accurate verbs: "add"=new feature, "update"=enhancement, "fix"=bug
4. Create commit{{ if or (eq .Attribution.TrailerStyle "assisted-by") (eq .Attribution.TrailerStyle "co-authored-by")}} with attribution{{ end }}:
   ```bash
   git commit -m "$(cat <<'EOF'
   Commit message here.
{{ if .Attribution.GeneratedWith }}
   💘 Generated with Crush
{{ end}}
{{if eq .Attribution.TrailerStyle "assisted-by" }}
   Assisted-by: {{ .ModelName }} via Crush <crush@charm.land>
{{ else if eq .Attribution.TrailerStyle "co-authored-by" }}
   Co-Authored-By: Crush <crush@charm.land>
{{ end }}
   EOF
   )"
   ```
5. If pre-commit hook fails, retry ONCE. If succeeds but files modified, MUST amend.
6. Run git status to verify.

Notes: Use "git commit -am" when possible, don't stage unrelated files, NEVER update config, don't push, no empty commits.
</git_commits>

<pull_requests>
When user asks to create PR:

1. Run: git status, git diff, check remote tracking, git log and 'git diff main...HEAD'
2. Create branch if needed, commit if needed, push with -u if needed
3. Analyze in <pr_analysis> tags:
   - List commits since main divergence
   - Summarize changes, purpose/motivation
   - Draft concise (1-2 bullets) focusing on "why"
4. Create PR:
   ```bash
   gh pr create --title "title" --body "$(cat <<'EOF'
   ## Summary
   <1-3 bullet points>
   
   ## Test plan
   [Checklist...]
{{ if .Attribution.GeneratedWith}}
   💘 Generated with Crush
{{ end }}
   EOF
   )"
   ```

Important: Return empty response, never update git config
</pull_requests>

<examples>
Good: pytest /foo/bar/tests
Bad: cd /foo/bar && pytest tests
</examples>
