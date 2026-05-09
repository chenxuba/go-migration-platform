# Business Domain Map

## Institution Education Domain

Important institution-side workflows include:

- Student lifecycle: intended students, enrolled students, student details.
- Course and class operations: course categories, course pages, one-to-one classes, group classes, timetables.
- Finance operations: sale orders, order details, tuition accounts, approvals.
- Assessment workflows: PEP-3, 儿心量表-II, assessment records, report interpretation, PDF/Word export.
- IEP workflows: generate IEP total plans from assessment records, save drafts, confirm plans, generate monthly and weekly execution plans, export Word documents.

## IEP Flow

Institution backend IEP generation currently lives around:

- `institution-admin/src/pages/teacher-center/components/generate-iep-modal.vue`
- `institution-admin/src/pages/teacher-center/components/iep-plan-adapters.ts`
- `institution-admin/src/api/edu-center/pep3-assessment.ts`
- `institution-admin/src/api/edu-center/erxin-assessment.ts`

Core concepts:

- Assessment adapters support PEP-3 and 儿心量表-II.
- IEP total plan has draft and confirmed states.
- Plan duration supports 3 months and 6 months.
- Execution plans are generated from the total plan as monthly plans and weekly plans.
- Word export exists for total plans and execution plans.
- AI generation may depend on assessment results, report interpretation, training records, and the IEP research library.

## Pad Assessment App

The pad app is Flutter-based and centers on assessment assistant workflows.

Useful locations:

- `assessment_pad_app/lib/main.dart`: home and shortcut entry points.
- `assessment_pad_app/lib/assessment_report_list_page.dart`: assessment report list and preview.
- `assessment_pad_app/lib/assessment_scale_category_page.dart`: assessment start/category page.
- `assessment_pad_app/designs/`: static design artifacts for pad pages.

Pad UI should match existing warm, calm, business-friendly visual language unless the user asks for a different direction.
