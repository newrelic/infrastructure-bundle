. "C:\newrelic\Program Files\New Relic\newrelic-infra\installer.ps1"


net start newrelic-infra

# Start-Process "C:\newrelic\Program Files/New Relic/newrelic-infra/newrelic-infra-service.exe"
# Start-Process "C:\newrelic\New Relic\newrelic-infra\newrelic-integrations\bin\nri-apache.exe"
# Start-Process "C:\newrelic\New Relic\newrelic-infra\newrelic-integrations\bin\nri-cassandra.exe"
# Start-Process "C:\newrelic\New Relic\newrelic-infra\newrelic-integrations\bin\nri-consul.exe"
# Start-Process "C:\newrelic\New Relic\newrelic-infra\newrelic-integrations\bin\nri-couchbase.exe"
# Start-Process "C:\newrelic\New Relic\newrelic-infra\newrelic-integrations\bin\nri-elasticsearch.exe"
# Start-Process "C:\newrelic\New Relic\newrelic-infra\newrelic-integrations\bin\nri-f5.exe"
# Start-Process "C:\newrelic\New Relic\newrelic-infra\newrelic-integrations\bin\nri-haproxy.exe"
# Start-Process "C:\newrelic\New Relic\newrelic-infra\newrelic-integrations\bin\nri-jmx.exe"
# Start-Process "C:\newrelic\New Relic\newrelic-infra\newrelic-integrations\bin\nri-kafka.exe"
# Start-Process "C:\newrelic\New Relic\newrelic-infra\newrelic-integrations\bin\nri-memcached.exe"
# Start-Process "C:\newrelic\New Relic\newrelic-infra\newrelic-integrations\bin\nri-mongodb.exe"
# Start-Process "C:\newrelic\New Relic\newrelic-infra\newrelic-integrations\bin\nri-mssql.exe"
# Start-Process "C:\newrelic\New Relic\newrelic-infra\newrelic-integrations\bin\nri-mysql.exe"
# Start-Process "C:\newrelic\New Relic\newrelic-infra\newrelic-integrations\bin\nri-nagios.exe"
# Start-Process "C:\newrelic\New Relic\newrelic-infra\newrelic-integrations\bin\nri-nginx.exe"
# Start-Process "C:\newrelic\New Relic\newrelic-infra\newrelic-integrations\bin\nri-postgresql.exe"
# Start-Process "C:\newrelic\New Relic\newrelic-infra\newrelic-integrations\bin\nri-rabbitmq.exe"
# Start-Process "C:\newrelic\New Relic\newrelic-infra\newrelic-integrations\bin\nri-redis.exe"



# Start-Process "C:\newrelic\Program Files\New relic\nrjmx\bin\nrjmx.bat"

# Start-Process "C:\newrelic\var\db\newrelic-infra\nri-discovery-kubernetes.exe"








while ($true) { Start-Sleep -Seconds 60}