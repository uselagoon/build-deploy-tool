## lagoon-build validate docker-compose

Verify docker-compose file for compatability with this tool

```
lagoon-build validate docker-compose [flags]
```

### Options

```
      --docker-compose string   The docker-compose.yml file to read. (default "docker-compose.yml")
  -h, --help                    help for docker-compose
```

### Options inherited from parent commands

```
  -a, --active-environment string          Name of the active environment if known
  -b, --branch string                      The name of the branch
  -d, --build-type string                  The type of build (branch, pullrequest, promote)
      --default-backup-schedule string     The default backup schedule to use
  -e, --environment-name string            The environment name to check
  -E, --environment-type string            The type of environment (development or production)
      --environment-variables string       The JSON payload for environment scope variables
  -A, --fastly-api-secret-prefix string    The fastly secret prefix to use (default "fastly-api-")
  -F, --fastly-cache-no-cache-id string    The fastly cache no cache service ID to use
  -f, --fastly-service-id string           The fastly service ID to use
      --ignore-missing-env-files           Ignore missing env_file files (true by default, subject to change). (default true)
      --ignore-non-string-key-errors       Ignore non-string-key docker-compose errors (true by default, subject to change). (default true)
      --images string                      JSON representation of service:image reference
  -L, --lagoon-version string              The lagoon version
  -l, --lagoon-yml string                  The .lagoon.yml file to read (default ".lagoon.yml")
      --lagoon-yml-override string         The .lagoon.yml override file to read for merging values into target lagoon.yml (default ".lagoon.override.yml")
  -M, --monitoring-config string           The monitoring contact config if known
  -m, --monitoring-status-page-id string   The monitoring status page ID if known
      --print-resulting-lagoonyml          Display the resulting, post merging, lagoon.yml file.
  -p, --project-name string                The project name
      --project-variables string           The JSON payload for project scope variables
  -B, --pullrequest-base-branch string     The pullrequest base branch
  -H, --pullrequest-head-branch string     The pullrequest head branch
  -P, --pullrequest-number string          The pullrequest number
      --pullrequest-title string           The pullrequest title
  -T, --saved-templates-path string        Path to where the resulting templates are saved (default "/kubectl-build-deploy/lagoon/services-routes")
  -s, --standby-environment string         Name of the standby environment if known
  -t, --template-path string               Path to the template on disk (default "/kubectl-build-deploy/")
```

### SEE ALSO

* [lagoon-build validate](lagoon-build_validate.md)	 - Validate resources

