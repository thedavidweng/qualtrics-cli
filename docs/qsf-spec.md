# Markdown Survey Specification

Reference for `qualtrics definitions build <spec.md>`, which compiles this
format to a `.qsf` file for **Create Project → Import a QSF File**.
No API access required.

```markdown
---
title: My Survey Title
language: EN
description: Optional survey description (metadata only)
---

# Block Name

## Question text [type]
- Choice 1
- Choice 2
```

## Frontmatter

YAML block at the top of the file.

| Key | Default | Description |
|-----|---------|-------------|
| `title` | `Untitled Survey` | Survey name shown in Qualtrics |
| `language` | `EN` | Language code |
| `description` | `""` | Survey description (metadata only, not shown to respondents) |
| `languages` | `[]` | Extra languages, e.g. `languages: ES, FR` (enables translations) |

## Blocks

Level-1 headings (`#`) define blocks. Without any `#` heading, all questions
go into a single default block.

## Questions

Level-2 headings (`##`) define questions, with the type in trailing brackets.
Append `*` to require an answer, and `@label` to name the question for
cross-references (`skip-if`, `show-if`, `branch-if`, `loop-from`,
`carry-from` resolve `@label` to the generated QID automatically):

```markdown
## What is your age? [mc]* @age
- Under 18
- 18-24
```

## Question types

| Type | Qualtrics mapping | Notes |
|------|-------------------|-------|
| `[mc]` | Multiple Choice, single answer | Choices as `- ` bullets |
| `[mc-multi]` | Multiple Choice, multiple answers | Choices as `- ` bullets |
| `[mc-dropdown]` | Multiple Choice, dropdown | Choices as `- ` bullets |
| `[rank]` | Rank order, drag and drop | Choices as `- ` bullets; supports `carry-from:` |
| `[text]` | Single-line text entry | No bullets |
| `[text-essay]` | Multi-line essay entry | No bullets |
| `[matrix]` | Matrix / Likert, one answer per row | `scale:` line + `- ` row bullets (not WCAG AA; prefer `[likert]`) |
| `[matrix-multi]` | Matrix, multiple answers per row | Same syntax as `[matrix]` |
| `[likert]` | Rating set → individual `[mc]` questions | Same syntax as `[matrix]`; accessible alternative |
| `[description]` | Descriptive text, no input | Heading is the label; body paragraphs are the display text |

An unknown `[type]` falls back to `[text]` with a build warning; any bullets
written for it are ignored.

Single-item Likert questions (e.g. "how likely would you be…") are written as
`[mc]` with the scale points as choices; there is no separate single-row type.

## Choices

```markdown
## What is your gender? [mc]*
- Man
- Woman
- Non-binary
- Prefer to self-describe [+text]
- Prefer not to say
```

- `[+text]` appends an inline text field to that choice (for "Other, please
  specify"). Plain `Other` without `[+text]` has no text field.
- `[exclusive]` marks the choice as exclusive: selecting it deselects all
  other choices (for "None of the above" / "No preference" in `[mc-multi]`).
  Combines with `[+text]` and `[VARNAME=N]`.
- `[VARNAME]` or `[VARNAME=N]` sets the export variable name and/or numeric
  recode value: `- Financial loss [FINANCE=9]`.

## Matrix

```markdown
## Rate the following features [matrix]
scale: Poor, Fair, Good, Excellent
- Ease of use
- Visual design
- Performance
```

The `scale:` line is comma-separated columns; bullets are rows. A matrix
without `scale:` or without rows builds but emits a warning and will import
as a broken question — fix the spec instead.

Likert matrices are **not** WCAG 2.0 AA compliant: Qualtrics flags them, warns
about mobile, and rejects them for Synthetic panels. For rating scales that
must pass the checker, use `[likert]`.

## Likert sets (accessible rating scales)

Same syntax as `[matrix]`, but each row becomes its own `[mc]` question at
build time (`QuestionText` is stem + row, choices are the scale points).
Generated export tags are `@label_1…N`, or per-row `[VARNAME]` when given:

```markdown
## How important is this feature to you? [likert]* @features
scale: Not at all important, Slightly important, Moderately important, Important, Very important
- Creating karaoke tracks from compatible personal music files
- Synchronized lyrics [LYRICS]
```

Every generated question is a plain multiple-choice question, so the set
passes WCAG checks, mobile tips, and Synthetic-panel compatibility. `@label`
resolves to the first generated question; logic referencing it warns.
`[exclusive]`, `[+text]`, recodes, `min/max-answers`, and `carry-from` have
no meaning on `[likert]` rows and warn.

## Answer-count limits

On `[mc-multi]` questions, `min-answers:` / `max-answers:` lines declare how
many choices respondents may select:

```markdown
## Select up to two. [mc-multi]*
max-answers: 2
- Official product website
- No preference [exclusive]
- Other [+text]
```

The limits are validated (positive, `min <= max`, `max <=` choice count) and
kept in the question text for respondents. They are **not** enforced by the
generated QSF: no public Qualtrics export, API schema, or reverse-engineering
note documents a storage shape for answer ranges, and shipping an unverified
shape risks silent import corruption. The build emits a warning with the
manual step: after import, set Response Requirements → answer range in the
Qualtrics editor (under a minute per question).

## Description blocks

```markdown
## Instructions [description]
Please read the following carefully before proceeding.

All responses are anonymous.
```

Body paragraphs become the displayed text (blank lines separate paragraphs).
If there is no body text, the heading itself is used as the display text so
the block never imports blank.

## Page breaks

A `---` line inside a block inserts a page break between questions.

## Labels

`@label` on a question heading names it for logic references; labels are
resolved to QIDs at build time. Duplicate labels and references to undefined
labels each emit a build warning (duplicate export tags corrupt analysis, so
keep labels unique).

## Logic

Placed on the line(s) immediately after the heading they belong to
(`branch-if:` / `loop-from:` go immediately after `# Block Name`, before its
first question):

```markdown
## Are you subscribed? [mc]*
- Yes
- No
skip-if: 1 Selected → ENDOFBLOCK
```

| Directive | Syntax | Values |
|-----------|--------|--------|
| `skip-if:` | `<choice#> <condition> → <destination>` | condition: `Selected`, `NotSelected`, `Empty`, `NotEmpty`; destination: `ENDOFBLOCK`, `ENDOFSURVEY`, `QID<n>` / `@label` |
| `show-if:` | `QID<n>/<choice#> <operator>` or `@label/<choice#> <operator>` | operator: `Selected`, `NotSelected`; on a question shows that question, on a block shows all its questions |
| `branch-if:` | same as `show-if:` | wraps the whole block in a Branch flow element |
| `loop-from:` | `QID<n>` or `@label` (must be `[mc-multi]`) | repeats the block per selected choice; use `${lm://Field/1}` in text for the current choice |
| `carry-from:` | `QID<n>` or `@label` (must be `[mc-multi]`) | fills the question's choices from another question's selections; no bullets needed |

## Translations

```markdown
## What is your age? [mc]*
lang-de: Wie alt sind Sie?
- Under 18
  lang-de: Unter 18 Jahren
```

- `lang-XX:` after the heading translates question text; after a bullet
  translates that choice/row.
- `lang-XX-scale:` with a comma-separated list translates the matrix scale.
- Requires the language in frontmatter `languages:`.

## Response requirements

`[type]*` writes Force Response (`ForceResponse` + `ForceResponseType` both
`ON`); without `*` the question is optional (`OFF`/`ON`). `[description]`
questions carry no response setting, matching real exports.

## Build warnings

`definitions build` prints warnings to stderr (stdout stays parseable) for:
unknown types, questions without choices, matrices without rows/scale,
duplicate or unresolved `@label`s, empty description blocks, misused
`[exclusive]` or `min/max-answers`, and answer-count limits (which need the
manual editor step below).
A clean spec builds with no warnings; treat any warning as a defect in the
spec, not noise.

`QuestionDescription` (the internal name shown in Qualtrics) is truncated to
99 characters; the respondent-facing `QuestionText` is never truncated.

## Payload fidelity

Question payloads mirror verified Qualtrics exports (Tufts, UBC, Utrecht,
Princeton, Auckland samples): `ForceResponseType`, string `"TextEntry"`,
boolean `ExclusiveAnswer` (including `false` on matrix rows), Type-only
`Validation` on `[description]`, and the `BL → FL → PROJ → QC → RS → SCO →
SO → SQ → STAT` element order. Fields whose meaning is unverified
(`AnalyzeChoices`, content-validation siblings, answer-range storage) are
deliberately not emitted.

## Limitations (set manually in Qualtrics after import)

- Answer-count limits ("select up to two", see above): kept as instruction
  text, enforced via Response Requirements in the editor after import.
- Branch conditions support a single `QID/choice operator` expression (no
  AND/OR); loop blocks cannot combine with `branch-if:`.
- Scoring, choice randomization, and survey-level options (back button,
  progress bar) are not configurable from the spec; survey options import
  with fixed defaults (public link, save-and-continue on, no progress bar).
