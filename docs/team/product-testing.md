# Product & Testing

Use this guide to turn business requirements into testable acceptance criteria and decide whether a feature is ready for demonstration or customer use.

## Developer handoff

For each feature, developers should provide:

- A short statement of the user problem and implemented behavior
- The branch or commit under test
- Setup steps and required test data
- Known limitations and areas not included
- Automated checks that passed
- The user flows most affected by the change

Product should reject a handoff that only says “done” without a testable behavior.

## Test environments

Use a disposable or designated test tenant. Do not test imports, destructive lifecycle actions, or customization releases against customer production data.

Prepare realistic roles, products, warehouses, prices, and opening stock. Keep the original input file so results can be compared.

## Acceptance checklist

Confirm:

- The main user journey completes without manual database or container intervention
- Validation messages explain what the user can correct
- Tenant data cannot appear in another tenant
- Permissions match the intended user role
- Refresh, retry, and failure states behave predictably
- Mobile and desktop layouts remain usable where applicable
- Existing critical workflows still work
- Documentation and demo instructions match the tested behavior

## Product import testing

Test CSV and XLSX files with machine code, model code, description, ETB and USD prices, pictures where supported, and stock columns. Cover valid rows, repeated machine codes, missing required values, invalid numbers, and a representative customer workbook before approval.

## Defect report

Include the tenant, build or commit, role, exact steps, expected result, actual result, evidence, frequency, and business impact. Mark a release blocked when it risks data loss, tenant isolation, authentication, provisioning, or a primary customer workflow.

## Release decision

Product approval means acceptance criteria passed and known limitations are acceptable. Engineering test success alone is not product acceptance; conversely, a visually successful demo does not replace security, migration, or automated checks.
