# Qualtrics CLI

A Go CLI that wraps the Qualtrics v3 REST API for humans and agents, covering the Survey Platform and XM Directory surface first, with later product lines tracked in the README roadmap.

## Language

### Account & access

**Brand**:
A tenant organization inside Qualtrics. Every API call is scoped to exactly one brand, and a user may belong to several.
_Avoid_: account, org, organization, tenant

**Division**:
A subdivision of a brand used to partition users and groups. Users and groups belong to divisions, not directly to the brand.
_Avoid_: department, team, org unit

**Datacenter**:
The Qualtrics region hosting a brand (e.g. `pdx1`, `iad1`, `fra1`). It prefixes the API base URL: `https://<datacenter>.qualtrics.com/API/v3/`.
_Avoid_: region, zone, endpoint

**API Token**:
A per-user secret sent in the `X-API-TOKEN` header, generated once under Account Settings → Qualtrics IDs. It authenticates as the user who created it.
_Avoid_: API key, password, bearer

### Surveys

**Survey**:
A questionnaire identified by an `SV_`-prefixed ID. It owns a definition, responses, and the distributions that collect them.
_Avoid_: questionnaire, form, project

**Survey Definition**:
The full structural content of a survey (questions, blocks, flow, answer choices) as a single nested JSON document, addressed at `/survey-definitions/{surveyId}`.
_Avoid_: survey JSON, schema, blueprint

**Survey Element**:
One typed piece of a survey definition, identified by an `Element` code: `BL` (blocks), `FL` (flow), `SO` (options), `SQ` (question). Elements are the unit of granular create/update/delete.
_Avoid_: component, node, section

**QSF**:
Qualtrics Survey Format — a JSON file wrapping a survey definition for import through the web UI. It has no public spec and no API import path; its structure is known only from reverse-engineered exports.
_Avoid_: survey file, export file

**Question**:
A single item inside a survey definition. Questions live in blocks; blocks live in the flow.
_Avoid_: item, field, prompt

**Block**:
A grouping of questions inside a survey definition that respondents see together.
_Avoid_: page, section, group

### Responses

**Response Export**:
An asynchronous job that packages survey responses into a downloadable file (CSV, JSON, SPSS, or TSV). Created with a format and a survey ID, polled by progress ID, then downloaded as a zip.
_Avoid_: export, download, report

**Progress**:
The state of an asynchronous job (`inProgress`, `complete`, `failed`) together with its percent complete, keyed by a progress ID.
_Avoid_: job status, task, poll result

**Response Import**:
An asynchronous job that uploads a file of responses into a survey.
_Avoid_: upload, ingest

**Behavioral Model**:
A declared set of priors (branch proportions, multi-select tendencies, scale means, directional hypotheses) used to generate synthetic responses for testing analysis pipelines. Roadmap material, not part of v1.
_Avoid_: fake data spec, mock responses

### Distribution & contacts

**Distribution**:
A send of a survey to an audience, identified by a `EMD_`-prefixed ID. It carries the links (individual, anonymous, or mailing-list) that respondents use.
_Avoid_: campaign, send, blast, email

**Distribution Link**:
A unique survey-taking URL belonging to a distribution. Links are anonymous, individual, or mailing-list flavored.
_Avoid_: URL, invite, link

**Directory**:
An XM Directory container for a brand's contacts, identified by a `POOL_`-prefixed ID. Directories contain mailing lists; they do not contain contacts directly.
_Avoid_: contact list, audience, pool

**Mailing List**:
A named group of contacts inside a directory, identified by a `CG_`-prefixed ID.
_Avoid_: list, segment, group (groups are a User Management concept)

**Contact**:
A person record in a mailing list, identified by a `MLR_`-prefixed ID, carrying embedded data and unsubscribe state.
_Avoid_: recipient, lead, subscriber

### Webhooks

**Event Subscription**:
A webhook registration that posts survey events (e.g. `surveyengine.partialResponse`, `surveyengine.completedResponse`) to a HTTPS URL.
_Avoid_: webhook, hook, callback
