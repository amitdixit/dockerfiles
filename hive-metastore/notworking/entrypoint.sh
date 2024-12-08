#!/bin/bash

# Use default values if environment variables are not set
POSTGRESQL_URL=${POSTGRESQL_URL:-"jdbc:postgresql://localhost:5432/metastore_db"}
POSTGRESQL_USER=${POSTGRESQL_USER:-"postgres"}
POSTGRESQL_PASSWORD=${POSTGRESQL_PASSWORD:-"Password@0"}

# ENV SERVICE_OPTS="-Djavax.jdo.option.ConnectionDriverName=org.postgresql.Driver \
#                   -Djavax.jdo.option.ConnectionURL=jdbc:postgresql://localhost:5432/custom_db \
#                   -Djavax.jdo.option.ConnectionUserName=postgres \
#                   -Djavax.jdo.option.ConnectionPassword=Password@0"

# Export SERVICE_OPTS for Hive Metastore
export SERVICE_OPTS="-Djavax.jdo.option.ConnectionDriverName=org.postgresql.Driver \
-Djavax.jdo.option.ConnectionURL=${POSTGRESQL_URL} \
-Djavax.jdo.option.ConnectionUserName=${POSTGRESQL_USER} \
-Djavax.jdo.option.ConnectionPassword=${POSTGRESQL_PASSWORD}"

schematool -initSchema -dbType postgres
 
# Debug output (optional, for troubleshooting)
if [ "$DEBUG" = "true" ]; then
  echo "SERVICE_OPTS: $SERVICE_OPTS"
fi
echo "SERVICE_OPTS: $SERVICE_OPTS"
# Execute the passed command or default to Hive Metastore
if [ "$#" -eq 0 ]; then
  exec hive --service metastore
else
  exec "$@"
fi
