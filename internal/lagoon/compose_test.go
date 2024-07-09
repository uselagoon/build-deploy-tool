package lagoon

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshaDockerComposeYAML(t *testing.T) {
	type args struct {
		file                     string
		ignoreNonStringKeyErrors bool
		ignoreMissingEnvFiles    bool
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
		want       string
		wantOrder  []OriginalServiceOrder
	}{
		{
			name: "test1 docker-compose drupal example",
			args: args{
				file: "../../test-resources/docker-compose/test1/docker-compose.yml",
			},
			want: `{"name":"test1","networks":{"amazeeio-network":{"name":"amazeeio-network","ipam":{},"external":true},"default":{"name":"test1_default","ipam":{}}},"services":{"cli":{"build":{"context":".","dockerfile":"lagoon/cli.dockerfile"},"command":null,"entrypoint":null,"environment":{"LAGOON_ROUTE":"http://test1.docker.amazee.io"},"image":"test1-cli","labels":{"lagoon.persistent":"/app/web/sites/default/files/","lagoon.persistent.name":"nginx","lagoon.type":"cli-persistent","lando.type":"php-cli-drupal"},"networks":{"default":null},"user":"root","volumes":[{"type":"bind","source":".","target":"/app","bind":{"create_host_path":true}},{"type":"volume","source":"files","target":"/app/web/sites/default/files","volume":{}}],"volumes_from":["container:amazeeio-ssh-agent"]},"mariadb":{"command":null,"entrypoint":null,"environment":{"LAGOON_ROUTE":"http://test1.docker.amazee.io"},"image":"uselagoon/mariadb-10.5-drupal:latest","labels":{"lagoon.type":"mariadb","lando.type":"mariadb-drupal"},"networks":{"default":null},"ports":[{"mode":"ingress","target":3306,"protocol":"tcp"}],"user":"1000"},"nginx":{"build":{"context":".","dockerfile":"lagoon/nginx.dockerfile","args":{"CLI_IMAGE":"test1-cli"}},"command":null,"depends_on":{"cli":{"condition":"service_started","required":true}},"entrypoint":null,"environment":{"LAGOON_LOCALDEV_URL":"http://test1.docker.amazee.io","LAGOON_ROUTE":"http://test1.docker.amazee.io"},"labels":{"lagoon.persistent":"/app/web/sites/default/files/","lagoon.type":"nginx-php-persistent","lando.type":"nginx-drupal"},"networks":{"amazeeio-network":null,"default":null},"user":"1000","volumes":[{"type":"bind","source":".","target":"/app","bind":{"create_host_path":true}},{"type":"volume","source":"files","target":"/app/web/sites/default/files","volume":{}}]},"php":{"build":{"context":".","dockerfile":"lagoon/php.dockerfile","args":{"CLI_IMAGE":"test1-cli"}},"command":null,"depends_on":{"cli":{"condition":"service_started","required":true}},"entrypoint":null,"environment":{"LAGOON_ROUTE":"http://test1.docker.amazee.io"},"labels":{"lagoon.name":"nginx","lagoon.persistent":"/app/web/sites/default/files/","lagoon.type":"nginx-php-persistent","lando.type":"php-fpm"},"networks":{"default":null},"user":"1000","volumes":[{"type":"bind","source":".","target":"/app","bind":{"create_host_path":true}},{"type":"volume","source":"files","target":"/app/web/sites/default/files","volume":{}}]},"redis":{"command":null,"entrypoint":null,"environment":{"LAGOON_ROUTE":"http://test1.docker.amazee.io"},"image":"uselagoon/redis-5:latest","labels":{"lagoon.type":"redis","lando.type":"redis"},"networks":{"default":null},"ports":[{"mode":"ingress","target":6379,"protocol":"tcp"}],"user":"1000"},"solr":{"command":null,"entrypoint":null,"environment":{"LAGOON_ROUTE":"http://test1.docker.amazee.io"},"image":"uselagoon/solr-7.7-drupal:latest","labels":{"lagoon.type":"solr","lando.type":"solr-drupal"},"networks":{"default":null},"ports":[{"mode":"ingress","target":8983,"protocol":"tcp"}]}},"volumes":{"files":{"name":"test1_files"}},"x-environment":{"LAGOON_ROUTE":"http://test1.docker.amazee.io"},"x-user":{"user":"1000"},"x-volumes":{"volumes":[".:/app:delegated","files:/app/web/sites/default/files"]}}`,
			wantOrder: []OriginalServiceOrder{
				{Index: 0, Name: "cli"},
				{Index: 1, Name: "nginx"},
				{Index: 2, Name: "php"},
				{Index: 3, Name: "mariadb"},
				{Index: 4, Name: "redis"},
				{Index: 5, Name: "solr"},
			},
		},
		{
			name: "test2 docker-compose node example",
			args: args{
				file: "../../test-resources/docker-compose/test2/docker-compose.yml",
			},
			want: `{"name":"test2","networks":{"amazeeio-network":{"name":"amazeeio-network","ipam":{},"external":true},"default":{"name":"test2_default","ipam":{}}},"services":{"node":{"build":{"context":".","dockerfile":"node.dockerfile"},"command":null,"entrypoint":null,"environment":{"LAGOON_LOCALDEV_HTTP_PORT":"3000","LAGOON_ROUTE":"http://node.docker.amazee.io"},"labels":{"lagoon.type":"node"},"networks":{"amazeeio-network":null,"default":null},"volumes":[{"type":"bind","source":".","target":"/app","bind":{"create_host_path":true}}]}}}`,
			wantOrder: []OriginalServiceOrder{
				{Index: 0, Name: "node"},
			},
		},
		{
			name: "test3 docker-compose complex",
			args: args{
				file: "../../test-resources/docker-compose/test3/docker-compose.yml",
			},
			want: `{"name":"test3","networks":{"amazeeio-network":{"name":"amazeeio-network","ipam":{},"external":true},"default":{"name":"test3_default","ipam":{}}},"services":{"cli":{"build":{"context":".","dockerfile":".lagoon/cli.dockerfile","args":{"DOCKER_CLI_IMAGE_URI":"","ENVIRONMENT_TYPE_ID":""}},"command":null,"container_name":"_cli","entrypoint":null,"environment":{"ENVIRONMENT_TYPE_ID":"","LAGOON_ENVIRONMENT_TYPE":"","LAGOON_PROJECT":"","LAGOON_ROUTE":"http://","PHP_MEMORY_LIMIT":"768M","XDEBUG_ENABLE":""},"labels":{"lagoon.persistent":"/app/docroot/sites/default/files/","lagoon.persistent.name":"nginx","lagoon.type":"cli-persistent"},"networks":{"default":null},"user":"root","volumes":[{"type":"bind","source":"./.lagoon/scripts/bash_prompts.rc","target":"/home/.bashrc","bind":{"create_host_path":true}},{"type":"bind","source":"./.lagoon/scripts/color_grid.sh","target":"/home/color_grid.sh","bind":{"create_host_path":true}}],"volumes_from":["container:amazeeio-ssh-agent"]},"mariadb":{"command":null,"container_name":"_db","entrypoint":null,"environment":{"ENVIRONMENT_TYPE_ID":"","LAGOON_ENVIRONMENT_TYPE":"","LAGOON_PROJECT":"","LAGOON_ROUTE":"http://","PHP_MEMORY_LIMIT":"768M","XDEBUG_ENABLE":""},"image":"amazeeio/mariadb-drupal","labels":{"lagoon.type":"mariadb"},"networks":{"default":null},"ports":[{"mode":"ingress","target":3306,"protocol":"tcp"}],"volumes":[{"type":"volume","source":"mysql","target":"/var/lib/mysql","volume":{}}]},"nginx":{"build":{"context":".","dockerfile":".lagoon/nginx.dockerfile","args":{"CLI_IMAGE":"","DOCKER_NGINX_IMAGE_URI":"","LAGOON_GIT_BRANCH":""}},"command":null,"container_name":"_nginx","depends_on":{"cli":{"condition":"service_started","required":true}},"entrypoint":null,"environment":{"ENVIRONMENT_TYPE_ID":"","LAGOON_ENVIRONMENT_TYPE":"","LAGOON_LOCALDEV_URL":"http://","LAGOON_PROJECT":"","LAGOON_ROUTE":"http://","PHP_MEMORY_LIMIT":"768M","XDEBUG_ENABLE":""},"labels":{"lagoon.name":"nginx","lagoon.persistent":"/app/docroot/sites/default/files/","lagoon.type":"nginx-php-persistent"},"networks":{"amazeeio-network":null,"default":null},"volumes":[{"type":"bind","source":"./.lagoon/nginx/nginx-http.conf","target":"/etc/nginx/conf.d/000-nginx-http.conf","bind":{"create_host_path":true}},{"type":"bind","source":"./.lagoon/nginx/app.conf","target":"/etc/nginx/conf.d/app.conf","bind":{"create_host_path":true}}]},"php":{"build":{"context":".","dockerfile":".lagoon/php.dockerfile","args":{"CLI_IMAGE":"","DOCKER_PHP_IMAGE_URI":""}},"command":null,"container_name":"_php","depends_on":{"cli":{"condition":"service_started","required":true}},"entrypoint":null,"environment":{"ENVIRONMENT_TYPE_ID":"","LAGOON_ENVIRONMENT_TYPE":"","LAGOON_PROJECT":"","LAGOON_ROUTE":"http://","PHP_MEMORY_LIMIT":"768M","XDEBUG_ENABLE":""},"labels":{"lagoon.deployment.servicetype":"php","lagoon.name":"nginx","lagoon.persistent":"/app/docroot/sites/default/files","lagoon.type":"nginx-php-persistent"},"networks":{"default":null}}},"volumes":{"app":{"name":"test3_app"},"mysql":{"name":"test3_mysql"},"solr7":{"name":"test3_solr7"}},"x-environment":{"ENVIRONMENT_TYPE_ID":"","LAGOON_ENVIRONMENT_TYPE":"","LAGOON_PROJECT":"","LAGOON_ROUTE":"http://","PHP_MEMORY_LIMIT":"768M","XDEBUG_ENABLE":""},"x-lagoon-project":"","x-site-branch":"","x-user":{"user":""}}`,
			wantOrder: []OriginalServiceOrder{
				{Index: 0, Name: "cli"},
				{Index: 1, Name: "nginx"},
				{Index: 2, Name: "php"},
				{Index: 3, Name: "mariadb"},
			},
		},
		{
			name: "test4 docker-compose complex",
			args: args{
				file: "../../test-resources/docker-compose/test4/docker-compose.yml",
			},
			want: `{"name":"test4","networks":{"amazeeio-network":{"name":"amazeeio-network","ipam":{},"external":true},"default":{"name":"test4_default","ipam":{}}},"services":{"chrome":{"command":null,"depends_on":{"cli":{"condition":"service_started","required":true}},"entrypoint":null,"environment":{"CKEDITOR_SCAYT_CUSTOMERID":"","CKEDITOR_SCAYT_SLANG":"","DB_ALIAS":"example.prod-left","DRUPAL_HASH_SALT":"fakehashsaltfakehashsaltfakehashsalt","DRUPAL_REFRESH_SEARCHAPI":"","EXAMPLE_IMAGE_VERSION":"latest","EXAMPLE_INGRESS_ENABLED":"","EXAMPLE_INGRESS_HEADER":"","EXAMPLE_INGRESS_PSK":"","EXAMPLE_KEY":"","GITHUB_TOKEN":"","LAGOON_ENVIRONMENT_TYPE":"local","LAGOON_LOCALDEV_URL":"http://mysite.docker.amazee.io","LAGOON_PROJECT":"mysite","LAGOON_ROUTE":"http://mysite.docker.amazee.io","PHP_MEMORY_LIMIT":"1024M","REDIS_CACHE_PREFIX":"tide_"},"image":"seleniarm/standalone-chromium:101.0","labels":{"lagoon.type":"none"},"networks":{"default":null},"shm_size":"1073741824","volumes":[{"type":"bind","source":".","target":"/app","bind":{"create_host_path":true}},{"type":"bind","source":"./docroot/sites/default/files","target":"/app/docroot/sites/default/files","bind":{"create_host_path":true}}]},"clamav":{"command":null,"entrypoint":null,"environment":{"CKEDITOR_SCAYT_CUSTOMERID":"","CKEDITOR_SCAYT_SLANG":"","DB_ALIAS":"example.prod-left","DRUPAL_HASH_SALT":"fakehashsaltfakehashsaltfakehashsalt","DRUPAL_REFRESH_SEARCHAPI":"","EXAMPLE_IMAGE_VERSION":"latest","EXAMPLE_INGRESS_ENABLED":"","EXAMPLE_INGRESS_HEADER":"","EXAMPLE_INGRESS_PSK":"","EXAMPLE_KEY":"","GITHUB_TOKEN":"","LAGOON_ENVIRONMENT_TYPE":"local","LAGOON_LOCALDEV_URL":"http://mysite.docker.amazee.io","LAGOON_PROJECT":"mysite","LAGOON_ROUTE":"http://mysite.docker.amazee.io","PHP_MEMORY_LIMIT":"1024M","REDIS_CACHE_PREFIX":"tide_"},"image":"clamav/example-clamav:4.x","labels":{"lagoon.type":"none"},"networks":{"default":null},"ports":[{"mode":"ingress","target":3310,"protocol":"tcp"}]},"cli":{"build":{"context":".","dockerfile":".docker/Dockerfile.cli","args":{"COMPOSER":"composer.json","EXAMPLE_IMAGE_VERSION":"4.x"}},"command":null,"entrypoint":null,"environment":{"CKEDITOR_SCAYT_CUSTOMERID":"","CKEDITOR_SCAYT_SLANG":"","DB_ALIAS":"example.prod-left","DRUPAL_HASH_SALT":"fakehashsaltfakehashsaltfakehashsalt","DRUPAL_REFRESH_SEARCHAPI":"","EXAMPLE_IMAGE_VERSION":"latest","EXAMPLE_INGRESS_ENABLED":"","EXAMPLE_INGRESS_HEADER":"","EXAMPLE_INGRESS_PSK":"","EXAMPLE_KEY":"","GITHUB_TOKEN":"","LAGOON_ENVIRONMENT_TYPE":"local","LAGOON_LOCALDEV_URL":"http://mysite.docker.amazee.io","LAGOON_PROJECT":"mysite","LAGOON_ROUTE":"http://mysite.docker.amazee.io","PHP_MEMORY_LIMIT":"1024M","REDIS_CACHE_PREFIX":"tide_"},"image":"mysite","labels":{"lagoon.persistent":"/app/docroot/sites/default/files/","lagoon.persistent.name":"nginx-php","lagoon.persistent.size":"50Gi","lagoon.type":"cli-persistent"},"networks":{"default":null},"volumes":[{"type":"bind","source":".","target":"/app","bind":{"create_host_path":true}},{"type":"bind","source":"./docroot/sites/default/files","target":"/app/docroot/sites/default/files","bind":{"create_host_path":true}}],"volumes_from":["container:amazeeio-ssh-agent"]},"elasticsearch":{"build":{"context":".","dockerfile":".docker/Dockerfile.elasticsearch","args":{"ES_TPL":"elasticsearch.yml"}},"command":null,"entrypoint":null,"labels":{"lagoon.type":"none"},"networks":{"default":null}},"mariadb":{"command":null,"entrypoint":null,"environment":{"CKEDITOR_SCAYT_CUSTOMERID":"","CKEDITOR_SCAYT_SLANG":"","DB_ALIAS":"example.prod-left","DRUPAL_HASH_SALT":"fakehashsaltfakehashsaltfakehashsalt","DRUPAL_REFRESH_SEARCHAPI":"","EXAMPLE_IMAGE_VERSION":"latest","EXAMPLE_INGRESS_ENABLED":"","EXAMPLE_INGRESS_HEADER":"","EXAMPLE_INGRESS_PSK":"","EXAMPLE_KEY":"","GITHUB_TOKEN":"","LAGOON_ENVIRONMENT_TYPE":"local","LAGOON_LOCALDEV_URL":"http://mysite.docker.amazee.io","LAGOON_PROJECT":"mysite","LAGOON_ROUTE":"http://mysite.docker.amazee.io","PHP_MEMORY_LIMIT":"1024M","REDIS_CACHE_PREFIX":"tide_"},"image":"uselagoon/mariadb-10.4-drupal:latest","labels":{"lagoon.type":"mariadb-shared"},"networks":{"default":null},"ports":[{"mode":"ingress","target":3306,"protocol":"tcp"}]},"nginx":{"build":{"context":".","dockerfile":".docker/Dockerfile.nginx-drupal","args":{"CLI_IMAGE":"mysite","EXAMPLE_IMAGE_VERSION":"4.x"}},"command":null,"depends_on":{"cli":{"condition":"service_started","required":true}},"entrypoint":null,"environment":{"CKEDITOR_SCAYT_CUSTOMERID":"","CKEDITOR_SCAYT_SLANG":"","DB_ALIAS":"example.prod-left","DRUPAL_HASH_SALT":"fakehashsaltfakehashsaltfakehashsalt","DRUPAL_REFRESH_SEARCHAPI":"","EXAMPLE_IMAGE_VERSION":"latest","EXAMPLE_INGRESS_ENABLED":"","EXAMPLE_INGRESS_HEADER":"","EXAMPLE_INGRESS_PSK":"","EXAMPLE_KEY":"","GITHUB_TOKEN":"","LAGOON_ENVIRONMENT_TYPE":"local","LAGOON_LOCALDEV_URL":"http://mysite.docker.amazee.io","LAGOON_PROJECT":"mysite","LAGOON_ROUTE":"http://mysite.docker.amazee.io","PHP_MEMORY_LIMIT":"1024M","REDIS_CACHE_PREFIX":"tide_"},"expose":["8080"],"labels":{"lagoon.name":"nginx-php","lagoon.persistent":"/app/docroot/sites/default/files/","lagoon.persistent.size":"50Gi","lagoon.type":"nginx-php-persistent"},"networks":{"amazeeio-network":null,"default":null},"volumes":[{"type":"bind","source":".","target":"/app","bind":{"create_host_path":true}},{"type":"bind","source":"./docroot/sites/default/files","target":"/app/docroot/sites/default/files","bind":{"create_host_path":true}}]},"php":{"build":{"context":".","dockerfile":".docker/Dockerfile.php","args":{"CLI_IMAGE":"mysite","EXAMPLE_IMAGE_VERSION":"4.x"}},"command":null,"depends_on":{"cli":{"condition":"service_started","required":true}},"entrypoint":null,"environment":{"CKEDITOR_SCAYT_CUSTOMERID":"","CKEDITOR_SCAYT_SLANG":"","DB_ALIAS":"example.prod-left","DRUPAL_HASH_SALT":"fakehashsaltfakehashsaltfakehashsalt","DRUPAL_REFRESH_SEARCHAPI":"","EXAMPLE_IMAGE_VERSION":"latest","EXAMPLE_INGRESS_ENABLED":"","EXAMPLE_INGRESS_HEADER":"","EXAMPLE_INGRESS_PSK":"","EXAMPLE_KEY":"","GITHUB_TOKEN":"","LAGOON_ENVIRONMENT_TYPE":"local","LAGOON_LOCALDEV_URL":"http://mysite.docker.amazee.io","LAGOON_PROJECT":"mysite","LAGOON_ROUTE":"http://mysite.docker.amazee.io","PHP_MEMORY_LIMIT":"1024M","REDIS_CACHE_PREFIX":"tide_"},"labels":{"lagoon.name":"nginx-php","lagoon.persistent":"/app/docroot/sites/default/files/","lagoon.persistent.size":"50Gi","lagoon.type":"nginx-php-persistent"},"networks":{"default":null},"volumes":[{"type":"bind","source":".","target":"/app","bind":{"create_host_path":true}},{"type":"bind","source":"./docroot/sites/default/files","target":"/app/docroot/sites/default/files","bind":{"create_host_path":true}}]},"redis":{"command":null,"entrypoint":null,"image":"uselagoon/redis-5:latest","labels":{"lagoon.type":"redis"},"networks":{"default":null}}},"volumes":{"app":{"name":"test4_app"},"files":{"name":"test4_files"}},"x-environment":{"CKEDITOR_SCAYT_CUSTOMERID":"","CKEDITOR_SCAYT_SLANG":"","DB_ALIAS":"example.prod-left","DRUPAL_HASH_SALT":"fakehashsaltfakehashsaltfakehashsalt","DRUPAL_REFRESH_SEARCHAPI":"","EXAMPLE_IMAGE_VERSION":"latest","EXAMPLE_INGRESS_ENABLED":"","EXAMPLE_INGRESS_HEADER":"","EXAMPLE_INGRESS_PSK":"","EXAMPLE_KEY":"","GITHUB_TOKEN":"","LAGOON_ENVIRONMENT_TYPE":"local","LAGOON_LOCALDEV_URL":"http://mysite.docker.amazee.io","LAGOON_PROJECT":"mysite","LAGOON_ROUTE":"http://mysite.docker.amazee.io","PHP_MEMORY_LIMIT":"1024M","REDIS_CACHE_PREFIX":"tide_"},"x-example-image-version":"4.x","x-project":"mysite","x-volumes":{"volumes":[".:/app:delegated","./docroot/sites/default/files:/app/docroot/sites/default/files:delegated"]}}`,
			wantOrder: []OriginalServiceOrder{
				{Index: 0, Name: "cli"},
				{Index: 1, Name: "nginx"},
				{Index: 2, Name: "php"},
				{Index: 3, Name: "mariadb"},
				{Index: 4, Name: "redis"},
				{Index: 5, Name: "elasticsearch"},
				{Index: 6, Name: "chrome"},
				{Index: 7, Name: "clamav"},
			},
		},
		{
			name: "test5 docker-compose complex",
			args: args{
				file: "../../test-resources/docker-compose/test5/docker-compose.yml",
			},
			want: `{"name":"test5","networks":{"amazeeio-network":{"name":"amazeeio-network","ipam":{},"external":true},"default":{"name":"test5_default","ipam":{}}},"services":{"chrome":{"command":null,"depends_on":{"cli":{"condition":"service_started","required":true}},"entrypoint":null,"environment":{"CI":"","DOCKERHOST":"host.docker.internal","LAGOON_ENVIRONMENT_TYPE":"local","LAGOON_LOCALDEV_URL":"example-project.docker.amazee.io","LAGOON_PROJECT":"example-project","LAGOON_ROUTE":"example-project.docker.amazee.io","PHP_APC_SHM_SIZE":"256M","PHP_MAX_EXECUTION_TIME":"-1","PHP_MAX_INPUT_VARS":"4000","PHP_MEMORY_LIMIT":"2G","XDEBUG_ENABLE":""},"image":"selenium/standalone-chrome:3.141.59-oxygen","labels":{"lagoon.type":"none"},"networks":{"default":null},"shm_size":"1073741824","volumes":[{"type":"bind","source":".","target":"/app","bind":{"create_host_path":true}}]},"cli":{"build":{"context":".","dockerfile":".docker/Dockerfile.cli"},"command":null,"entrypoint":null,"environment":{"CI":"","DOCKERHOST":"host.docker.internal","LAGOON_ENVIRONMENT_TYPE":"local","LAGOON_LOCALDEV_URL":"example-project.docker.amazee.io","LAGOON_PROJECT":"example-project","LAGOON_ROUTE":"example-project.docker.amazee.io","PHP_APC_SHM_SIZE":"256M","PHP_MAX_EXECUTION_TIME":"-1","PHP_MAX_INPUT_VARS":"4000","PHP_MEMORY_LIMIT":"2G","XDEBUG_ENABLE":""},"image":"example-project","labels":{"lagoon.persistent":"/app/docroot/sites/default/files/","lagoon.persistent.name":"nginx-php","lagoon.type":"cli-persistent"},"networks":{"default":null},"ports":[{"mode":"ingress","target":35729,"protocol":"tcp"}],"user":"root","volumes":[{"type":"bind","source":".","target":"/app","bind":{"create_host_path":true}}],"volumes_from":["container:amazeeio-ssh-agent"]},"mariadb":{"build":{"context":".","dockerfile":".docker/Dockerfile.mariadb","args":{"IMAGE":"amazeeio/mariadb-drupal:21.7.0"}},"command":null,"entrypoint":null,"environment":{"CI":"","DOCKERHOST":"host.docker.internal","LAGOON_ENVIRONMENT_TYPE":"local","LAGOON_LOCALDEV_URL":"example-project.docker.amazee.io","LAGOON_PROJECT":"example-project","LAGOON_ROUTE":"example-project.docker.amazee.io","PHP_APC_SHM_SIZE":"256M","PHP_MAX_EXECUTION_TIME":"-1","PHP_MAX_INPUT_VARS":"4000","PHP_MEMORY_LIMIT":"2G","XDEBUG_ENABLE":""},"labels":{"lagoon.type":"mariadb"},"networks":{"default":null},"ports":[{"mode":"ingress","target":3306,"protocol":"tcp"}]},"nginx":{"build":{"context":".","dockerfile":".docker/Dockerfile.nginx-drupal","args":{"CLI_IMAGE":"example-project"}},"command":null,"depends_on":{"cli":{"condition":"service_started","required":true}},"entrypoint":null,"environment":{"CI":"","DOCKERHOST":"host.docker.internal","LAGOON_ENVIRONMENT_TYPE":"local","LAGOON_LOCALDEV_URL":"example-project.docker.amazee.io","LAGOON_PROJECT":"example-project","LAGOON_ROUTE":"example-project.docker.amazee.io","PHP_APC_SHM_SIZE":"256M","PHP_MAX_EXECUTION_TIME":"-1","PHP_MAX_INPUT_VARS":"4000","PHP_MEMORY_LIMIT":"2G","XDEBUG_ENABLE":""},"labels":{"lagoon.name":"nginx-php","lagoon.persistent":"/app/docroot/sites/default/files/","lagoon.persistent.class":"slow","lagoon.type":"nginx-php-persistent"},"networks":{"amazeeio-network":null,"default":null},"user":"1000","volumes":[{"type":"bind","source":".","target":"/app","bind":{"create_host_path":true}}]},"php":{"build":{"context":".","dockerfile":".docker/Dockerfile.php","args":{"CLI_IMAGE":"example-project"}},"command":null,"depends_on":{"cli":{"condition":"service_started","required":true}},"entrypoint":null,"environment":{"CI":"","DOCKERHOST":"host.docker.internal","LAGOON_ENVIRONMENT_TYPE":"local","LAGOON_LOCALDEV_URL":"example-project.docker.amazee.io","LAGOON_PROJECT":"example-project","LAGOON_ROUTE":"example-project.docker.amazee.io","PHP_APC_SHM_SIZE":"256M","PHP_MAX_EXECUTION_TIME":"-1","PHP_MAX_INPUT_VARS":"4000","PHP_MEMORY_LIMIT":"2G","XDEBUG_ENABLE":""},"labels":{"lagoon.name":"nginx-php","lagoon.persistent":"/app/docroot/sites/default/files/","lagoon.persistent.class":"slow","lagoon.type":"nginx-php-persistent"},"networks":{"default":null},"user":"1000","volumes":[{"type":"bind","source":".","target":"/app","bind":{"create_host_path":true}}]},"redis":{"command":null,"entrypoint":null,"environment":{"CI":"","DOCKERHOST":"host.docker.internal","LAGOON_ENVIRONMENT_TYPE":"local","LAGOON_LOCALDEV_URL":"example-project.docker.amazee.io","LAGOON_PROJECT":"example-project","LAGOON_ROUTE":"example-project.docker.amazee.io","PHP_APC_SHM_SIZE":"256M","PHP_MAX_EXECUTION_TIME":"-1","PHP_MAX_INPUT_VARS":"4000","PHP_MEMORY_LIMIT":"2G","XDEBUG_ENABLE":""},"image":"amazeeio/redis:6-21.11.0","labels":{"lagoon.type":"redis"},"networks":{"default":null}},"wait_dependencies":{"command":["mariadb:3306"],"depends_on":{"cli":{"condition":"service_started","required":true},"mariadb":{"condition":"service_started","required":true}},"entrypoint":null,"image":"dadarek/wait-for-dependencies","labels":{"lagoon.type":"none"},"networks":{"default":null}}},"volumes":{"app":{"name":"test5_app"}},"x-environment":{"CI":"","DOCKERHOST":"host.docker.internal","LAGOON_ENVIRONMENT_TYPE":"local","LAGOON_LOCALDEV_URL":"example-project.docker.amazee.io","LAGOON_PROJECT":"example-project","LAGOON_ROUTE":"example-project.docker.amazee.io","PHP_APC_SHM_SIZE":"256M","PHP_MAX_EXECUTION_TIME":-1,"PHP_MAX_INPUT_VARS":4000,"PHP_MEMORY_LIMIT":"2G","XDEBUG_ENABLE":""},"x-project":"example-project","x-user":{"user":"1000"},"x-volumes":{"volumes":[".:/app:delegated"]}}`,
			wantOrder: []OriginalServiceOrder{
				{Index: 0, Name: "cli"},
				{Index: 1, Name: "nginx"},
				{Index: 2, Name: "php"},
				{Index: 3, Name: "mariadb"},
				{Index: 4, Name: "redis"},
				{Index: 5, Name: "chrome"},
				{Index: 6, Name: "wait_dependencies"},
			},
		},
		{
			name: "test6 docker-compose complex",
			args: args{
				file: "../../test-resources/docker-compose/test6/docker-compose.yml",
			},
			want: `{"name":"test6","networks":{"amazeeio-network":{"name":"amazeeio-network","ipam":{},"external":true},"default":{"name":"test6_default","ipam":{}}},"services":{"chrome":{"command":null,"depends_on":{"test":{"condition":"service_started","required":true}},"entrypoint":null,"image":"selenium/standalone-chrome","labels":{"lagoon.type":"none"},"networks":{"default":null},"shm_size":"1073741824","volumes":[{"type":"bind","source":"./themes","target":"/app/web/themes/custom","bind":{"create_host_path":true}},{"type":"bind","source":"./files","target":"/app/web/sites/default/files","bind":{"create_host_path":true}},{"type":"bind","source":"./tests/behat/features","target":"/app/tests/behat/features","bind":{"create_host_path":true}},{"type":"bind","source":"./tests/behat/screenshots","target":"/app/tests/behat/screenshots","bind":{"create_host_path":true}},{"type":"bind","source":"./tests/phpunit/tests","target":"/app/tests/phpunit/tests","bind":{"create_host_path":true}},{"type":"bind","source":"./config","target":"/app/config","bind":{"create_host_path":true}}]},"cli":{"build":{"context":".","dockerfile":".docker/Dockerfile.cli","args":{"EXAMPLE_IMAGE_VERSION":"9.x-latest","LAGOON_SAFE_PROJECT":"ca-learning2"}},"command":null,"entrypoint":null,"environment":{"DEV_MODE":"false","DOCKERHOST":"host.docker.internal","DRUPAL_SHIELD_PASS":"","DRUPAL_SHIELD_USER":"","EXAMPLE_DEPLOY_WORKFLOW_CONFIG":"import","EXAMPLE_IMAGE_VERSION":"9.x-latest","EXAMPLE_PREPARE_XML_SCRIPT":"/app/vendor/bin/example-prepare-xml","LAGOON_ENVIRONMENT_TYPE":"","LAGOON_PROJECT":"ca-learning2","LAGOON_ROUTE":"http://ca-learning2.docker.amazee.io","STAGE_FILE_PROXY_URL":"","XDEBUG_ENABLE":"","X_FRAME_OPTIONS":"SameOrigin"},"image":"ca-learning2","labels":{"lagoon.persistent":"/app/web/sites/default/files/","lagoon.persistent.name":"nginx","lagoon.type":"cli-persistent"},"networks":{"default":null},"volumes":[{"type":"bind","source":"./themes","target":"/app/web/themes/custom","bind":{"create_host_path":true}},{"type":"bind","source":"./files","target":"/app/web/sites/default/files","bind":{"create_host_path":true}},{"type":"bind","source":"./tests/behat/features","target":"/app/tests/behat/features","bind":{"create_host_path":true}},{"type":"bind","source":"./tests/behat/screenshots","target":"/app/tests/behat/screenshots","bind":{"create_host_path":true}},{"type":"bind","source":"./tests/phpunit/tests","target":"/app/tests/phpunit/tests","bind":{"create_host_path":true}},{"type":"bind","source":"./config","target":"/app/config","bind":{"create_host_path":true}}],"volumes_from":["container:amazeeio-ssh-agent"]},"mariadb":{"command":null,"entrypoint":null,"environment":{"DEV_MODE":"false","DOCKERHOST":"host.docker.internal","DRUPAL_SHIELD_PASS":"","DRUPAL_SHIELD_USER":"","EXAMPLE_DEPLOY_WORKFLOW_CONFIG":"import","EXAMPLE_IMAGE_VERSION":"9.x-latest","EXAMPLE_PREPARE_XML_SCRIPT":"/app/vendor/bin/example-prepare-xml","LAGOON_ENVIRONMENT_TYPE":"","LAGOON_PROJECT":"ca-learning2","LAGOON_ROUTE":"http://ca-learning2.docker.amazee.io","STAGE_FILE_PROXY_URL":"","XDEBUG_ENABLE":"","X_FRAME_OPTIONS":"SameOrigin"},"image":"example/mariadb-drupal:9.x-latest","labels":{"lagoon.image":"example/mariadb-drupal:9.x-latest","lagoon.type":"mariadb"},"networks":{"default":null},"ports":[{"mode":"ingress","target":3306,"protocol":"tcp"}]},"nginx":{"build":{"context":".","dockerfile":".docker/Dockerfile.nginx-drupal","args":{"CLI_IMAGE":"ca-learning2","EXAMPLE_IMAGE_VERSION":"9.x-latest"}},"command":null,"depends_on":{"cli":{"condition":"service_started","required":true}},"entrypoint":null,"environment":{"DEV_MODE":"false","DOCKERHOST":"host.docker.internal","DRUPAL_SHIELD_PASS":"","DRUPAL_SHIELD_USER":"","EXAMPLE_DEPLOY_WORKFLOW_CONFIG":"import","EXAMPLE_IMAGE_VERSION":"9.x-latest","EXAMPLE_PREPARE_XML_SCRIPT":"/app/vendor/bin/example-prepare-xml","LAGOON_ENVIRONMENT_TYPE":"","LAGOON_LOCALDEV_URL":"http://ca-learning2.docker.amazee.io","LAGOON_PROJECT":"ca-learning2","LAGOON_ROUTE":"http://ca-learning2.docker.amazee.io","STAGE_FILE_PROXY_URL":"","XDEBUG_ENABLE":"","X_FRAME_OPTIONS":"SameOrigin"},"labels":{"lagoon.persistent":"/app/web/sites/default/files/","lagoon.type":"nginx-php-persistent"},"networks":{"amazeeio-network":null,"default":null},"volumes":[{"type":"bind","source":"./themes","target":"/app/web/themes/custom","bind":{"create_host_path":true}},{"type":"bind","source":"./files","target":"/app/web/sites/default/files","bind":{"create_host_path":true}},{"type":"bind","source":"./tests/behat/features","target":"/app/tests/behat/features","bind":{"create_host_path":true}},{"type":"bind","source":"./tests/behat/screenshots","target":"/app/tests/behat/screenshots","bind":{"create_host_path":true}},{"type":"bind","source":"./tests/phpunit/tests","target":"/app/tests/phpunit/tests","bind":{"create_host_path":true}},{"type":"bind","source":"./config","target":"/app/config","bind":{"create_host_path":true}}]},"php":{"build":{"context":".","dockerfile":".docker/Dockerfile.php","args":{"CLI_IMAGE":"ca-learning2","EXAMPLE_IMAGE_VERSION":"9.x-latest"}},"command":null,"depends_on":{"cli":{"condition":"service_started","required":true}},"entrypoint":null,"environment":{"DEV_MODE":"false","DOCKERHOST":"host.docker.internal","DRUPAL_SHIELD_PASS":"","DRUPAL_SHIELD_USER":"","EXAMPLE_DEPLOY_WORKFLOW_CONFIG":"import","EXAMPLE_IMAGE_VERSION":"9.x-latest","EXAMPLE_PREPARE_XML_SCRIPT":"/app/vendor/bin/example-prepare-xml","LAGOON_ENVIRONMENT_TYPE":"","LAGOON_PROJECT":"ca-learning2","LAGOON_ROUTE":"http://ca-learning2.docker.amazee.io","STAGE_FILE_PROXY_URL":"","XDEBUG_ENABLE":"","X_FRAME_OPTIONS":"SameOrigin"},"labels":{"lagoon.name":"nginx","lagoon.persistent":"/app/web/sites/default/files/","lagoon.type":"nginx-php-persistent"},"networks":{"default":null},"volumes":[{"type":"bind","source":"./themes","target":"/app/web/themes/custom","bind":{"create_host_path":true}},{"type":"bind","source":"./files","target":"/app/web/sites/default/files","bind":{"create_host_path":true}},{"type":"bind","source":"./tests/behat/features","target":"/app/tests/behat/features","bind":{"create_host_path":true}},{"type":"bind","source":"./tests/behat/screenshots","target":"/app/tests/behat/screenshots","bind":{"create_host_path":true}},{"type":"bind","source":"./tests/phpunit/tests","target":"/app/tests/phpunit/tests","bind":{"create_host_path":true}},{"type":"bind","source":"./config","target":"/app/config","bind":{"create_host_path":true}}]},"test":{"build":{"context":".","dockerfile":".docker/Dockerfile.test","args":{"CLI_IMAGE":"ca-learning2","EXAMPLE_IMAGE_VERSION":"9.x-latest","SITE_AUDIT_VERSION":"7.x-3.x"}},"command":null,"depends_on":{"cli":{"condition":"service_started","required":true}},"entrypoint":null,"environment":{"DEV_MODE":"false","DOCKERHOST":"host.docker.internal","DRUPAL_SHIELD_PASS":"","DRUPAL_SHIELD_USER":"","EXAMPLE_DEPLOY_WORKFLOW_CONFIG":"import","EXAMPLE_IMAGE_VERSION":"9.x-latest","EXAMPLE_PREPARE_XML_SCRIPT":"/app/vendor/bin/example-prepare-xml","LAGOON_ENVIRONMENT_TYPE":"","LAGOON_PROJECT":"ca-learning2","LAGOON_ROUTE":"http://ca-learning2.docker.amazee.io","STAGE_FILE_PROXY_URL":"","XDEBUG_ENABLE":"","X_FRAME_OPTIONS":"SameOrigin"},"labels":{"lagoon.type":"none"},"networks":{"default":null},"volumes":[{"type":"bind","source":"./themes","target":"/app/web/themes/custom","bind":{"create_host_path":true}},{"type":"bind","source":"./files","target":"/app/web/sites/default/files","bind":{"create_host_path":true}},{"type":"bind","source":"./tests/behat/features","target":"/app/tests/behat/features","bind":{"create_host_path":true}},{"type":"bind","source":"./tests/behat/screenshots","target":"/app/tests/behat/screenshots","bind":{"create_host_path":true}},{"type":"bind","source":"./tests/phpunit/tests","target":"/app/tests/phpunit/tests","bind":{"create_host_path":true}},{"type":"bind","source":"./config","target":"/app/config","bind":{"create_host_path":true}}]}},"x-environment":{"DEV_MODE":"false","DOCKERHOST":"host.docker.internal","DRUPAL_SHIELD_PASS":"","DRUPAL_SHIELD_USER":"","EXAMPLE_DEPLOY_WORKFLOW_CONFIG":"import","EXAMPLE_IMAGE_VERSION":"9.x-latest","EXAMPLE_PREPARE_XML_SCRIPT":"/app/vendor/bin/example-prepare-xml","LAGOON_ENVIRONMENT_TYPE":"","LAGOON_PROJECT":"ca-learning2","LAGOON_ROUTE":"http://ca-learning2.docker.amazee.io","STAGE_FILE_PROXY_URL":"","XDEBUG_ENABLE":"","X_FRAME_OPTIONS":"SameOrigin"},"x-example-image-version":"9.x-latest","x-lagoon-local-dev-url":"http://ca-learning2.docker.amazee.io","x-lagoon-project":"ca-learning2","x-volumes":{"volumes":["./themes:/app/web/themes/custom:delegated","./files:/app/web/sites/default/files:delegated","./tests/behat/features:/app/tests/behat/features:delegated","./tests/behat/screenshots:/app/tests/behat/screenshots:delegated","./tests/phpunit/tests:/app/tests/phpunit/tests:delegated","./config:/app/config"]},"x-volumes-paas":{"volumes":[".:/app:delegated"]}}`,
			wantOrder: []OriginalServiceOrder{
				{Index: 0, Name: "cli"},
				{Index: 1, Name: "test"},
				{Index: 2, Name: "nginx"},
				{Index: 3, Name: "php"},
				{Index: 4, Name: "mariadb"},
				{Index: 5, Name: "chrome"},
			},
		},
		{
			name: "test7 check an invalid docker-compose with ignoring non-string key errors",
			args: args{
				file:                     "../../test-resources/docker-compose/test7/docker-compose.yml",
				ignoreNonStringKeyErrors: true,
			},
			want: `{"name":"test7","networks":{"amazeeio-network":{"name":"amazeeio-network","ipam":{},"external":true},"default":{"name":"test7_default","ipam":{}}},"services":{"cli":{"build":{"context":".","dockerfile":".lagoon/cli.dockerfile","args":{"DOCKER_CLI_IMAGE_URI":"","ENVIRONMENT_TYPE_ID":""}},"command":null,"container_name":"_cli","entrypoint":null,"environment":{"ENVIRONMENT_TYPE_ID":"","LAGOON_ENVIRONMENT_TYPE":"","LAGOON_PROJECT":"","LAGOON_ROUTE":"http://","PHP_MEMORY_LIMIT":"768M","XDEBUG_ENABLE":""},"labels":{"lagoon.persistent":"/app/docroot/sites/default/files/","lagoon.persistent.name":"nginx","lagoon.type":"cli-persistent"},"networks":{"default":null},"user":"root","volumes":[{"type":"bind","source":"./.lagoon/scripts/bash_prompts.rc","target":"/home/.bashrc","bind":{"create_host_path":true}},{"type":"bind","source":"./.lagoon/scripts/color_grid.sh","target":"/home/color_grid.sh","bind":{"create_host_path":true}}],"volumes_from":["container:amazeeio-ssh-agent"]},"mariadb":{"command":null,"container_name":"_db","entrypoint":null,"environment":{"ENVIRONMENT_TYPE_ID":"","LAGOON_ENVIRONMENT_TYPE":"","LAGOON_PROJECT":"","LAGOON_ROUTE":"http://","PHP_MEMORY_LIMIT":"768M","XDEBUG_ENABLE":""},"image":"amazeeio/mariadb-drupal","labels":{"lagoon.type":"mariadb"},"networks":{"default":null},"ports":[{"mode":"ingress","target":3306,"protocol":"tcp"}],"volumes":[{"type":"volume","source":"mysql","target":"/var/lib/mysql","volume":{}}]},"nginx":{"build":{"context":".","dockerfile":".lagoon/nginx.dockerfile","args":{"CLI_IMAGE":"","DOCKER_NGINX_IMAGE_URI":""}},"command":null,"container_name":"_nginx","depends_on":{"cli":{"condition":"service_started","required":true}},"entrypoint":null,"environment":{"ENVIRONMENT_TYPE_ID":"","LAGOON_ENVIRONMENT_TYPE":"","LAGOON_LOCALDEV_URL":"http://","LAGOON_PROJECT":"","LAGOON_ROUTE":"http://","PHP_MEMORY_LIMIT":"768M","XDEBUG_ENABLE":""},"labels":{"lagoon.name":"nginx","lagoon.persistent":"/app/docroot/sites/default/files/","lagoon.type":"nginx-php-persistent"},"networks":{"amazeeio-network":null,"default":null},"volumes":[{"type":"bind","source":"./.lagoon/nginx/nginx-http.conf","target":"/etc/nginx/conf.d/000-nginx-http.conf","bind":{"create_host_path":true}},{"type":"bind","source":"./.lagoon/nginx/app.conf","target":"/etc/nginx/conf.d/app.conf","bind":{"create_host_path":true}}]},"php":{"build":{"context":".","dockerfile":".lagoon/php.dockerfile","args":{"CLI_IMAGE":"","DOCKER_PHP_IMAGE_URI":""}},"command":null,"container_name":"_php","depends_on":{"cli":{"condition":"service_started","required":true}},"entrypoint":null,"environment":{"ENVIRONMENT_TYPE_ID":"","LAGOON_ENVIRONMENT_TYPE":"","LAGOON_PROJECT":"","LAGOON_ROUTE":"http://","PHP_MEMORY_LIMIT":"768M","XDEBUG_ENABLE":""},"labels":{"lagoon.deployment.servicetype":"php","lagoon.name":"nginx","lagoon.persistent":"/app/docroot/sites/default/files","lagoon.type":"nginx-php-persistent"},"networks":{"default":null}}},"volumes":{"app":{"name":"test7_app"},"mysql":{"name":"test7_mysql"},"solr7":{"name":"test7_solr7"}},"x-environment":{"ENVIRONMENT_TYPE_ID":"","LAGOON_ENVIRONMENT_TYPE":"","LAGOON_PROJECT":"","LAGOON_ROUTE":"http://","PHP_MEMORY_LIMIT":"768M","XDEBUG_ENABLE":""},"x-lagoon-project":"","x-site-branch":{},"x-user":{"user":""}}`,
			wantOrder: []OriginalServiceOrder{
				{Index: 0, Name: "cli"},
				{Index: 1, Name: "nginx"},
				{Index: 2, Name: "php"},
				{Index: 3, Name: "mariadb"},
			},
		},
		{
			name: "test8 check an invalid docker-compose (same as test7 but not ignoring the errors)",
			args: args{
				file: "../../test-resources/docker-compose/test8/docker-compose.yml",
			},
			wantErr:    true,
			wantErrMsg: "Non-string key in x-site-branch: <nil>",
		},
		{
			name: "test9 check an valid docker-compose with missing env_files",
			args: args{
				file:                     "../../test-resources/docker-compose/test9/docker-compose.yml",
				ignoreNonStringKeyErrors: true,
				ignoreMissingEnvFiles:    true,
			},
			want: `{"name":"test9","networks":{"default":{"name":"test9_default","ipam":{}},"stonehenge-network":{"name":"stonehenge-network","ipam":{},"external":true}},"services":{"cli":{"build":{"context":".","dockerfile":"lagoon/cli.dockerfile"},"command":null,"container_name":"test9-cli","entrypoint":null,"environment":{"DRUSH_OPTIONS_URI":"https://","LAGOON_PROJECT":"test9","LAGOON_ROUTE":"https://","SIMPLETEST_BASE_URL":"http://nginx:8080","SIMPLETEST_DB":"mysql://drupal:drupal@mariadb:3306/drupal","SSMTP_MAILHUB":"host.docker.internal:1025"},"image":"test9","labels":{"lagoon.persistent":"/app/public/sites/default/files/","lagoon.persistent.name":"nginx","lagoon.type":"cli-persistent"},"networks":{"default":null},"volumes":[{"type":"bind","source":".","target":"/app","bind":{"create_host_path":true}},{"type":"volume","source":"ssh","target":"/tmp/amazeeio_ssh-agent","volume":{}}]},"mariadb":{"command":null,"container_name":"test9-db","entrypoint":null,"environment":{"LAGOON_PROJECT":"test9","LAGOON_ROUTE":"https://","SSMTP_MAILHUB":"host.docker.internal:1025"},"image":"uselagoon/mariadb-drupal:latest","labels":{"lagoon.type":"mariadb"},"networks":{"default":null},"ports":[{"mode":"ingress","target":3306,"protocol":"tcp"}]},"nginx":{"build":{"context":".","dockerfile":"lagoon/nginx.dockerfile","args":{"CLI_IMAGE":"test9"}},"command":null,"container_name":"test9-nginx","depends_on":{"cli":{"condition":"service_started","required":true}},"entrypoint":null,"environment":{"LAGOON_LOCALDEV_URL":"","LAGOON_PROJECT":"test9","LAGOON_ROUTE":"https://","SSMTP_MAILHUB":"host.docker.internal:1025"},"labels":{"lagoon.persistent":"/app/public/sites/default/files/","lagoon.type":"nginx-php-persistent"},"networks":{"default":null,"stonehenge-network":null},"volumes":[{"type":"bind","source":".","target":"/app","bind":{"create_host_path":true}}]},"php":{"build":{"context":".","dockerfile":"lagoon/php.dockerfile","args":{"CLI_IMAGE":"test9"}},"command":null,"container_name":"test9-php","depends_on":{"cli":{"condition":"service_started","required":true}},"entrypoint":null,"environment":{"LAGOON_PROJECT":"test9","LAGOON_ROUTE":"https://","SSMTP_MAILHUB":"host.docker.internal:1025"},"labels":{"lagoon.name":"nginx","lagoon.persistent":"/app/public/sites/default/files/","lagoon.type":"nginx-php-persistent"},"networks":{"default":null},"volumes":[{"type":"bind","source":".","target":"/app","bind":{"create_host_path":true}}]},"pma":{"command":null,"container_name":"test9-pma","entrypoint":null,"environment":{"PMA_HOST":"mariadb","PMA_PASSWORD":"drupal","PMA_USER":"drupal","UPLOAD_LIMIT":"1G"},"image":"phpmyadmin/phpmyadmin","labels":{"lagoon.type":"none"},"networks":{"default":null,"stonehenge-network":null}}},"volumes":{"es_data":{"name":"test9_es_data"},"ssh":{"name":"stonehenge-ssh","external":true}},"x-environment":{"LAGOON_PROJECT":"test9","LAGOON_ROUTE":"https://","SSMTP_MAILHUB":"host.docker.internal:1025"},"x-lagoon-project":"test9"}`,
			wantOrder: []OriginalServiceOrder{
				{Index: 0, Name: "cli"},
				{Index: 1, Name: "nginx"},
				{Index: 2, Name: "php"},
				{Index: 3, Name: "mariadb"},
				{Index: 4, Name: "pma"},
			},
		},
		{
			name: "test10 check an valid docker-compose with missing env_files (same as test9 but not ignoring the errors)",
			args: args{
				file: "../../test-resources/docker-compose/test10/docker-compose.yml",
			},
			wantErr:    true,
			wantErrMsg: "no such file or directory",
		},
		{
			name: "test11 docker-compose service name with '.'",
			args: args{
				file: "../../test-resources/docker-compose/test11/docker-compose.yml",
			},
			wantErr:    true,
			wantErrMsg: "Service name is invalid. Please refer to the documentation regarding service naming requirements",
		},
		{
			name: "test12 check a valid docker-compose with newer version of optional env_files",
			args: args{
				file:                     "../../test-resources/docker-compose/test12/docker-compose.yml",
				ignoreNonStringKeyErrors: true,
				ignoreMissingEnvFiles:    true,
			},
			want: `{"name":"test12","networks":{"default":{"name":"test12_default","ipam":{}},"stonehenge-network":{"name":"stonehenge-network","ipam":{},"external":true}},"services":{"cli":{"build":{"context":".","dockerfile":"lagoon/cli.dockerfile"},"command":null,"container_name":"test12-cli","entrypoint":null,"environment":{"DRUSH_OPTIONS_URI":"https://","LAGOON_PROJECT":"test12","LAGOON_ROUTE":"https://","SIMPLETEST_BASE_URL":"http://nginx:8080","SIMPLETEST_DB":"mysql://drupal:drupal@mariadb:3306/drupal","SSMTP_MAILHUB":"host.docker.internal:1025"},"image":"test12","labels":{"lagoon.persistent":"/app/public/sites/default/files/","lagoon.persistent.name":"nginx","lagoon.type":"cli-persistent"},"networks":{"default":null},"volumes":[{"type":"bind","source":".","target":"/app","bind":{"create_host_path":true}},{"type":"volume","source":"ssh","target":"/tmp/amazeeio_ssh-agent","volume":{}}]},"mariadb":{"command":null,"container_name":"test12-db","entrypoint":null,"environment":{"LAGOON_PROJECT":"test12","LAGOON_ROUTE":"https://","SSMTP_MAILHUB":"host.docker.internal:1025"},"image":"uselagoon/mariadb-drupal:latest","labels":{"lagoon.type":"mariadb"},"networks":{"default":null},"ports":[{"mode":"ingress","target":3306,"protocol":"tcp"}]},"nginx":{"build":{"context":".","dockerfile":"lagoon/nginx.dockerfile","args":{"CLI_IMAGE":"test12"}},"command":null,"container_name":"test12-nginx","depends_on":{"cli":{"condition":"service_started","required":true}},"entrypoint":null,"environment":{"LAGOON_LOCALDEV_URL":"","LAGOON_PROJECT":"test12","LAGOON_ROUTE":"https://","SSMTP_MAILHUB":"host.docker.internal:1025"},"labels":{"lagoon.persistent":"/app/public/sites/default/files/","lagoon.type":"nginx-php-persistent"},"networks":{"default":null,"stonehenge-network":null},"volumes":[{"type":"bind","source":".","target":"/app","bind":{"create_host_path":true}}]},"php":{"build":{"context":".","dockerfile":"lagoon/php.dockerfile","args":{"CLI_IMAGE":"test12"}},"command":null,"container_name":"test12-php","depends_on":{"cli":{"condition":"service_started","required":true}},"entrypoint":null,"environment":{"LAGOON_PROJECT":"test12","LAGOON_ROUTE":"https://","SSMTP_MAILHUB":"host.docker.internal:1025"},"labels":{"lagoon.name":"nginx","lagoon.persistent":"/app/public/sites/default/files/","lagoon.type":"nginx-php-persistent"},"networks":{"default":null},"volumes":[{"type":"bind","source":".","target":"/app","bind":{"create_host_path":true}}]},"pma":{"command":null,"container_name":"test12-pma","entrypoint":null,"environment":{"PMA_HOST":"mariadb","PMA_PASSWORD":"drupal","PMA_USER":"drupal","UPLOAD_LIMIT":"1G"},"image":"phpmyadmin/phpmyadmin","labels":{"lagoon.type":"none"},"networks":{"default":null,"stonehenge-network":null}}},"volumes":{"es_data":{"name":"test12_es_data"},"ssh":{"name":"stonehenge-ssh","external":true}},"x-environment":{"LAGOON_PROJECT":"test12","LAGOON_ROUTE":"https://","SSMTP_MAILHUB":"host.docker.internal:1025"},"x-lagoon-project":"test12"}`,
			wantOrder: []OriginalServiceOrder{
				{Index: 0, Name: "cli"},
				{Index: 1, Name: "nginx"},
				{Index: 2, Name: "php"},
				{Index: 3, Name: "mariadb"},
				{Index: 4, Name: "pma"},
			},
		},
		{
			name: "test13 check a valid docker-compose with newer version of optional env_files (same as test12 but not ignoring the errors)",
			args: args{
				file: "../../test-resources/docker-compose/test13/docker-compose.yml",
			},
			wantErr:    true,
			wantErrMsg: "no such file or directory",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, dcpo, err := UnmarshaDockerComposeYAML(tt.args.file, tt.args.ignoreNonStringKeyErrors, tt.args.ignoreMissingEnvFiles, map[string]string{})
			if err != nil && !tt.wantErr {
				t.Errorf("UnmarshaDockerComposeYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("UnmarshaDockerComposeYAML() error = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				}
			} else {
				stra, _ := json.Marshal(l)
				if !cmp.Equal(string(stra), tt.want) {
					t.Errorf("UnmarshaDockerComposeYAML() = %v, want %v", string(stra), tt.want)
				}
			}
			if !cmp.Equal(dcpo, tt.wantOrder) {
				t.Errorf("UnmarshaDockerComposeYAML() = %v, want %v", dcpo, tt.wantOrder)
			}
		})
	}
}

func TestCheckLagoonLabel(t *testing.T) {
	type args struct {
		labels map[string]string
		label  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				labels: map[string]string{
					"lagoon.type":            "cli-persistent",
					"lagoon.persistent":      "/app/web/sites/default/files/",
					"lagoon.persistent.name": "nginx",
				},
				label: "lagoon.persistent.name",
			},
			want: "nginx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckServiceLagoonLabel(tt.args.labels, tt.args.label); got != tt.want {
				t.Errorf("CheckServiceLagoonLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnmarshalLagoonDockerComposeYAML(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name       string
		args       args
		wantErrMsg string
		wantErr    bool
	}{
		// {
		// 	name: "test1 docker-compose drupal example",
		// 	args: args{
		// 		file: "../../test-resources/docker-compose/test1/docker-compose.yml",
		// 	},
		// 	wantErr:    true,
		// 	wantErrMsg: `line 59: mapping key "<<" already defined at line 58`,
		// },
		{
			name: "test2 docker-compose node example",
			args: args{
				file: "../../test-resources/docker-compose/test2/docker-compose.yml",
			},
		},
		{
			name: "test3 docker-compose complex",
			args: args{
				file: "../../test-resources/docker-compose/test3/docker-compose.yml",
			},
		},
		{
			name: "test4 docker-compose complex",
			args: args{
				file: "../../test-resources/docker-compose/test4/docker-compose.yml",
			},
		},
		// {
		// 	name: "test5 docker-compose complex",
		// 	args: args{
		// 		file: "../../test-resources/docker-compose/test5/docker-compose.yml",
		// 	},
		// 	wantErr:    true,
		// 	wantErrMsg: `line 57: mapping key "<<" already defined at line 56`,
		// },
		{
			name: "test6 docker-compose complex",
			args: args{
				file: "../../test-resources/docker-compose/test6/docker-compose.yml",
			},
		},
		// these tests are specific to docker-compose validations, but will pass yaml validations
		{
			name: "test7 check an invalid docker-compose with ignoring non-string key errors (valid yaml)",
			args: args{
				file: "../../test-resources/docker-compose/test7/docker-compose.yml",
			},
		},
		{
			name: "test8 check an invalid docker-compose (same as test7 but not ignoring the errors)",
			args: args{
				file: "../../test-resources/docker-compose/test8/docker-compose.yml",
			},
		},
		{
			name: "test9 check an valid docker-compose with missing env_files",
			args: args{
				file: "../../test-resources/docker-compose/test9/docker-compose.yml",
			},
		},
		{
			name: "test10 check an valid docker-compose with missing env_files (same as test9 but not ignoring the errors)",
			args: args{
				file: "../../test-resources/docker-compose/test10/docker-compose.yml",
			},
		},
		{
			name: "test11 docker-compose service name with '.'",
			args: args{
				file: "../../test-resources/docker-compose/test11/docker-compose.yml",
			},
		},
		// ^^ these tests are specific to docker-compose validations, but will pass yaml validations
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUnmarshalDockerComposeYAML(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUnmarshalDockerComposeYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("ValidateUnmarshalDockerComposeYAML() error = %v, wantErr %v", err.Error(), tt.wantErrMsg)
				}
				return
			}
		})
	}
}
