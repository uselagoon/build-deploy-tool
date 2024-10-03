# Build Requirements

Lagoon uses the following information injected into the build pod, and retrieved from files within the git repository to determine how an environment is built.

## Required Files
* `.lagoon.yml`
* `docker-compose.yml`

### `.lagoon.yml`
See the docs [here](https://docs.lagoon.sh/using-lagoon-the-basics/lagoon-yml/)

### `docker-compose.yml`
See the docs [here](https://docs.lagoon.sh/using-lagoon-the-basics/docker-compose-yml/)

## Variables

These are variables that are injected into a build pod by `remote-controller`, some are provided by Lagoon core when a build is created, some are injected into the build from `remote-controller`

### Core provided

#### Main Variables
* `BUILD_TYPE` can be one of `branch|pullrequest|promote`
* `PROJECT` is the safed version of the project name
* `ENVIRONMENT` is the safed version of the environment name
* `BRANCH` is the unedited name of the branch
* `ENVIRONMENT_TYPE` can be one of `development|production`
* `ACTIVE_ENVIRONMENT` is populated with the current active environment if active/standby is enabled
* `STANDBY_ENVIRONMENT` is populated with the current standby environment if active/standby is enabled

#### Pullrequest Variables
* `PR_TITLE`
* `PR_NUMBER`
* `PR_HEAD_BRANCH`
* `PR_HEAD_SHA`
* `PR_BASE_BRANCH`
* `PR_BASE_SHA`

####  Promotion Variables
* `PROMOTION_SOURCE_ENVIRONMENT` contains the source environment name if this is a promotion type build

#### Environment Variables
* `LAGOON_PROJECT_VARIABLES` contains any project specific environment variables
* `LAGOON_ENVIRONMENT_VARIABLES` contains any environment specific environment variables

### Monitoring Variables
* `MONITORING_ALERTCONTACT`
* `MONITORING_STATUSPAGEID`

#### Build Variables
* `SOURCE_REPOSITORY` is the git repository
* `GIT_REF` is the git reference / commit
* `SUBFOLDER` if the project is in a subfolder, this variable contains the directory to change to
* `PROJECT_SECRET` is used for backups
* `KUBERNETES` is the kubernetes cluster name from Lagoon
* `REGISTRY` is the registry that is passed from Lagoon (will be deprecated)

### Remote provided

#### General variables
These are variables that can influence parts of a build

* `LAGOON_FASTLY_NOCACHE_SERVICE_ID` is a default cache no cache service id that can be consumed
* `NATIVE_CRON_POD_MINIMUM_FREQUENCY` changes the interval of which cronjobs go from inside cli pods to native k8s cronjobs (default 15m)

### Build Flags
The following are flags provided by `remote-controller` and used to influence build, these also have counterpart variables that omit the `FORCE|DEFAULT` from them that can be used inside of environment variables, `FORCE` flags cannot be overridden.

* `LAGOON_FEATURE_FLAG_FORCE_ROOTLESS_WORKLOAD`
* `LAGOON_FEATURE_FLAG_DEFAULT_ROOTLESS_WORKLOAD`
* `LAGOON_FEATURE_FLAG_FORCE_ISOLATION_NETWORK_POLICY`
* `LAGOON_FEATURE_FLAG_DEFAULT_ISOLATION_NETWORK_POLICY`
* `LAGOON_FEATURE_FLAG_FORCE_INSIGHTS`
* `LAGOON_FEATURE_FLAG_DEFAULT_INSIGHTS`
* `LAGOON_FEATURE_FLAG_FORCE_RWX_TO_RWO`
* `LAGOON_FEATURE_FLAG_DEFAULT_RWX_TO_RWO`

### Proxy related variables
If proxy has been enabled in `remote-controller`, then these variables will be injected to the buildpod to enabled proxy support

* `HTTP_PROXY / http_proxy`
* `HTTPS_PROXY / https_proxy`
* `NO_PROXY / no_proxy`

### API and Remote provided

### Backup related variables
These are all variables that are provided by either core or remote 
* `DEFAULT_BACKUP_SCHEDULE`
* `MONTHLY_BACKUP_DEFAULT_RETENTION`
* `WEEKLY_BACKUP_DEFAULT_RETENTION`
* `DAILY_BACKUP_DEFAULT_RETENTION`
* `HOURLY_BACKUP_DEFAULT_RETENTION`
* `LAGOON_FEATURE_BACKUP_PROD_SCHEDULE` (remote) / `LAGOON_BACKUP_PROD_SCHEDULE` (API)
* `LAGOON_FEATURE_BACKUP_DEV_SCHEDULE` (remote) / `LAGOON_BACKUP_DEV_SCHEDULE` (API)
* `LAGOON_FEATURE_BACKUP_PR_SCHEDULE` (remote) / `LAGOON_BACKUP_PR_SCHEDULE` (API)
* `K8UP_WEEKLY_RANDOM_FEATURE_FLAG`
