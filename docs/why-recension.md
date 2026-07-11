# Why Recension?

In textual criticism (philology), a **recension** is the process of establishing a trusted reading of a work by comparing multiple manuscript copies, identifying variants, and authorizing one form as the reference text.

Recension (this project) does the same for software behavior:

1. The SDK **records** how a workflow behaves for each test case (checks, assumptions, metrics / timings).
2. The server **collates** submissions across versions (batches) within a **team** (multi-tenant RBAC).
3. Your team **promotes** a trusted batch as the baseline — the living recension (admin+).
4. Later versions are compared against that baseline; unintended side-effects — including slower timings — appear as clearly as corrupted readings in a manuscript.

This is not unit testing with hard-coded expected values. It is behavioral comparison against an authorized prior version.

See the root [`README.md`](../README.md) for stack, roles, and OpenAPI; the dashboard element view shows metric duration bars vs baseline.
