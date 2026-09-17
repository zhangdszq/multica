SHUZ-152 production deployment

The feature is deployed on the existing v0.4.44 production database. Do not
restore the isolated test database over production: it contains disabled
integrations, detached runtimes and synthetic test users/projects.

The active deployment is `/opt/multica-production-shuz152/compose.json`, Compose
project `multica-production-shuz152`. This file contains the original production
environment and is private (0600). Do not commit or publish it.

Services:

- `multica-production-shuz152-backend`: original 127.0.0.1:8080 mapping.
- `multica-production-shuz152-frontend`: original 127.0.0.1:3000 mapping.
- Existing `multica-postgres-1`, network `multica_default` and upload volume
  `multica_backend_uploads` are reused without copying stale test data.
- Existing nginx, HTTPS domain and login/JWT/mail configuration remain in use.

The original `multica-backend-1` and `multica-frontend-1` containers remain
stopped, with their original images and configuration. They were not deleted.
Do not run the old Compose project alongside the active deployment: they use
the same ports and service aliases.

Start/update the active deployment:

```sh
docker compose -p multica-production-shuz152 \
  -f /opt/multica-production-shuz152/compose.json up -d
```

Rollback to the preserved original application:

```sh
sh /opt/multica-production-shuz152/rollback.sh
```

Rollback stops the new application containers and starts the old ones. It keeps
the current database and uploads, including writes made after deployment. The
new project columns are additive and compatible with the old application.
Do not restore the pre-migration dump during an ordinary application rollback.

The deployment's `private/` directory contains the verified pre-migration
database archive, upload archive, original configuration/container metadata,
configuration fingerprints, and confirmed project-creator attribution. The
cutover and verification logs record the migration and configuration checks.

Only permission fields for existing production projects were promoted from the
test copy. Test-only projects/users were excluded. Historical creator claims
were carried over only when explicitly confirmed; unknown creators remain null.
