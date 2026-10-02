1.  **Bypasses `isSafeSelect` validation** - Executes `tool.Action.Config.Query` via `ExecSQLRead` without calling the allow-list `isSafeSelect()` defined in the original file (`selectOnlyRe`/`dangerousRe`). This allows `INSERT/UPDATE/DELETE/DROP/ALTER`, comments, and multi-statements if the underlying DB executor is not strictly read-only.
2.  **Non-deterministic parameter binding** - `for _, v := range args` iterates a Go `map` in random order, so `$1,$2` bind to wrong values causing logic bugs and potential injection.
3.  **Silent panic suppression** - `defer func(){ if r:=recover(); r!=nil {} }()` hides programming errors and prevents debugging/DoS detection.
4.  **No input validation** - `Query` is taken directly from DB `ActionDef` without allow-listing or schema validation.

Do not apply the edit as provided and keep the original file unchanged.

If a generic executor is required, use a hardened version like this in `business/generic_executor.go` that preserves the original validation and deterministic binding:

