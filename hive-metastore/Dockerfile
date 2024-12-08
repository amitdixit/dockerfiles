FROM apache/hive:4.0.1

# Download and add the PostgreSQL JDBC driver to the Hive lib directory
COPY postgresql-42.7.4.jar /opt/hive/lib

# Set environment variables for Hive Metastore
ENV SERVICE_NAME=metastore
ENV DB_DRIVER=postgres
ENV SERVICE_OPTS="-Djavax.jdo.option.ConnectionDriverName=org.postgresql.Driver \
                  -Djavax.jdo.option.ConnectionURL=jdbc:postgresql://localhost:5432/custom_db \
                  -Djavax.jdo.option.ConnectionUserName=postgres \
                  -Djavax.jdo.option.ConnectionPassword=Password@0"

# Default command to run Hive Metastore
CMD ["hive", "--service", "metastore"]
