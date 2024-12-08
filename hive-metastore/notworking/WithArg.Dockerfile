FROM apache/hive:4.0.1

# Copy the PostgreSQL JDBC driver to the Hive lib directory
COPY postgresql-42.7.4.jar /opt/hive/lib/
COPY entrypoint.sh /entrypoint.sh

USER root
# Ensure the entrypoint script is executable
RUN chmod +x /entrypoint.sh

# Default values for PostgreSQL connection settings
ENV POSTGRESQL_URL="jdbc:postgresql://localhost:5432/metastore_db"
ENV POSTGRESQL_USER="postgres"
ENV POSTGRESQL_PASSWORD="Password@0"
ENV DB_DRIVER="org.postgresql.Driver"
ENV SERVICE_NAME="metastore"

ENV SERVICE_OPTS="-Djavax.jdo.option.ConnectionDriverName=org.postgresql.Driver \
    -Djavax.jdo.option.ConnectionURL=${POSTGRESQL_URL} \
    -Djavax.jdo.option.ConnectionUserName=${POSTGRESQL_USER} \
    -Djavax.jdo.option.ConnectionPassword=${POSTGRESQL_PASSWORD}"

# Set entrypoint to run the custom script
ENTRYPOINT ["/entrypoint.sh"]
CMD ["hive", "--service", "metastore"]
