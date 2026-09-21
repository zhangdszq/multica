Project access test deployment (SHUZ-152), based on v0.4.44.

The test environment must use a separate PostgreSQL container and volume. Never
point this build's DATABASE_URL at the production database. Migration
535_project_access adds project creator and access policy columns. Existing
projects keep workspace-wide access and an unknown creator because older
versions did not record project creators. Workspace ownership does not grant
project-creator permissions. This migration originally shipped as 500 in the
fork; its idempotent DDL supports both fresh and already-upgraded installations.
Migration 536 removes the incorrect historical
owner attribution from the initial test deployment, while preserving actual
creators of new projects. Historical creators must be verified separately;
the system does not guess them from project leads or resource uploaders.

Verify the creator migration on a disposable PostgreSQL database:

```sh
psql -X "$TEST_DATABASE_URL" -f server/internal/migrations/testdata/project_access_creators.sql
```

The regression uses temporary tables and rolls back. It checks fresh installs,
repair of historical attribution, idempotency, and preservation of new-project
creators and access policies. Migration 536 stops for restricted historical
projects so their creator can be verified before removing access management.

Project creators can select members in the project sidebar's Access section.
Enable “Only selected members”, select members, and save. The creator cannot lock
themselves out. Unselected members, including workspace administrators, do not
receive the project in lists or search; direct project, issue, resource and pin
lookups return 404 so the project cannot be discovered by title or known ID.
Access is checked in project writes, resources, issue reads/writes and query
surfaces, attachment download authorization, and websocket delivery. Existing
signed download capabilities retain their original expiration semantics.

The isolated server copy uses:

- Directory: `/opt/multica-shuz152`
- Docker network: `multica-shuz152-internal` (internal, no internet egress)
- PostgreSQL: `multica-shuz152-postgres`, volume `multica-shuz152-pgdata`
- Backend: `multica-shuz152-backend`
- Frontend: `multica-shuz152-frontend`
- Gateway: independent systemd service `multica-shuz152-gateway`, host port 3100
- Backend/frontend stay on the internal Docker network, with no published ports.
  The gateway resolves their private IPs at startup using `render-gateway.sh`.

Before starting the backend, pause copied autopilots and their triggers, revoke
channel installations, remove copied access/daemon/task tokens, cancel unfinished
tasks, and detach agents from production runtimes. Use a fresh JWT signing key.
Keep the database dump and all env/credential files mode 0600. The gateway opens directly to the Multica login page, without additional Basic
authentication. The application uses a test verification code with signup disabled. Do not copy production integration or mail credentials.

Build the backend and migrator with CGO_ENABLED=0 GOOS=linux GOARCH=amd64. Build
web with STANDALONE=true. Runtime images inherit the matching v0.4.44 images to
retain their Linux native dependencies. The frontend uses REMOTE_API_URL at
runtime. The gateway proxies /api, /auth, /ws, /v1, /uploads and /health to backend.

Stop only the test environment:

```sh
systemctl stop multica-shuz152-gateway
docker stop multica-shuz152-frontend multica-shuz152-backend multica-shuz152-postgres
```

Install `nginx.conf` as `nginx.conf.template` alongside `render-gateway.sh`.
Start the containers again, then start or restart the gateway service to resolve
their current private IPs. The gateway uses its own config and PID file; never reload or restart the production nginx service. Do not use the production compose project name
`multica`. Removing test containers does not remove the test database volume.
After testing, remove the dedicated security-group rule for port 3100 and the
test deployment only. Preserve the source patch and any test data still needed.

When rebasing this feature onto a newer upstream revision, assign an unused
migration number and review any newly introduced project/issue query surface.
