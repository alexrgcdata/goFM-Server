# goFM agent operating rules

These instructions apply only to the standalone `goFM-server` project.
OpenBridge is a separate project and must not be edited, moved, rebuilt, or
deployed from this workspace.

## Ownership and architecture

- The primary agent owns architecture, security decisions, final edits, and
  verification.
- Keep goFM independent from OpenBridge. A reverse-proxy URL does not authorize
  copying Go files into the OpenBridge frontend.
- Never place API keys, bearer tokens, FileMaker credentials, session tokens,
  encrypted vault keys, or production configuration in source control.
- Never accept caller-provided upstream hosts, executable names, PHP files, or
  operating-system commands.

## Token-conservation policy

- Default to one primary agent. Do not spawn subagents for small, sequential,
  tightly coupled, or context-heavy work.
- Spawn a subagent only when its task is independent, bounded, and cheaper than
  carrying the work in the primary context.
- Use at most two subagents concurrently. Do not permit nested subagents.
- Prefer one read-only scout and, only when justified, one focused reviewer.
- Never assign two agents to inspect the same files or answer the same question.
- Do not let multiple agents edit overlapping files. The primary agent applies
  and reconciles final changes.
- Give each subagent the smallest file list and context needed. Do not send the
  full conversation or repository unless unavoidable.
- Require concise results: findings, evidence, file references, and a clear
  pass/fail conclusion. Do not return raw logs, full file contents, or long
  narratives.
- Stop a worker when it repeats the parent investigation, expands scope, or
  cannot make progress.

## Model routing

DeepInfra models are used only through a configured and callable OpenClaw
runtime, CLI, MCP server, or other explicit bridge. Do not assume that a
DeepInfra API key makes those models available as native Codex subagents.

When an OpenClaw DeepInfra bridge is available, prefer these roles:

1. `deepinfra/deepseek-ai/DeepSeek-V4-Flash`
   - Default low-cost scout.
   - Use for file discovery, focused searches, log triage, test-output summaries,
     documentation extraction, and simple checklists.
   - Read-only; concise output.
2. `deepinfra/nvidia/NVIDIA-Nemotron-3-Super-120B-A12B`
   - Low-cost reviewer.
   - Use for bounded correctness, test-gap, or maintainability reviews after the
     primary agent identifies the changed files.
   - Read-only; report only actionable findings.
3. `deepinfra/moonshotai/Kimi-K2.7-Code`
   - Coding specialist and escalation model.
   - Use only for a difficult, isolated coding problem where the cheaper models
     are inadequate. Do not use for broad repository scans or routine tests.

If the OpenClaw bridge is unavailable, use native Codex subagents only when the
task still satisfies the conservation rules. Prefer a low-reasoning lightweight
model for scouts and reserve stronger models for concrete security or
correctness reviews.

## Subagent task contract

Every delegated task must specify:

- exact objective and allowed files;
- read-only or write authority;
- commands it may run;
- explicit exclusions;
- maximum expected output (normally 300-600 words);
- required evidence and completion condition.

Recommended worker prompt shape:

```text
Read only these files: <paths>.
Answer only: <bounded question>.
Do not edit, browse unrelated files, or repeat repository-wide scans.
Return at most 500 words with file references, actionable findings, and a
pass/fail conclusion. If there are no findings, say so directly.
```

## Verification

- Run focused tests after each logical security or adapter change.
- Before handoff, run `go test ./...`, `go vet ./...`, and build the applicable
  Windows or Linux binary.
- Use mock FileMaker services for automated tests. Never require real production
  credentials in tests.
- Report what was actually verified separately from what still requires a live
  FileMaker Server, hosting control panel, reverse proxy, or external account.

## User-facing build and upload instructions

- Always state explicitly whether a step is for React or Go.
- For React, always give the exact local source folder, the exact `pnpm run
  build` command, and the exact public server destination.
- For Go, always state that there is no `dist` folder and name the compiled
  executable path and the private server destination requirement.
- Never tell the owner to upload a folder without naming its contents and the
  resulting server tree.
- Never use placeholder paths such as `/home/YOUR_USERNAME/` as if they are
  real paths. Ask for the host-provided application path or clearly label the
  path as an example.
- Do not create ZIP deployment packages unless the owner explicitly asks for
  one. Prefer a clearly named upload folder with the exact files to copy.
