-- Working code as an explicit choice (G8).
--
-- WHY a column: §9 made the implementation phase author design sections and
-- retired the application-code path deliberately. A workflow whose deliverable IS
-- code — in the existing-codebase environment the patch to the client's repository
-- is the whole point — had no way to ask for it. This column is that request.
--
-- Default FALSE, which is exactly the behaviour every workflow had before this
-- migration: nothing starts producing code just because the column exists.
ALTER TABLE workflows
    ADD COLUMN IF NOT EXISTS deliver_code BOOLEAN NOT NULL DEFAULT FALSE;
