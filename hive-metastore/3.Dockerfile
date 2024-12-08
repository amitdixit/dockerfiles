FROM apache/hive:4.0.1

# Download and add the PostgreSQL JDBC driver to the Hive lib directory
COPY postgresql-42.7.4.jar /opt/hive/lib/

# Create a startup script
COPY start-metastore.sh /start-metastore.sh
USER root
RUN chmod +x /start-metastore.sh

# Set default environment variables
ENV METASTORE_DB_DRIVER=org.postgresql.Driver
ENV METASTORE_DB_URL=jdbc:postgresql://localhost:5432/metastore
ENV METASTORE_DB_USER=postgres
ENV METASTORE_DB_PASSWORD=postgres

# Entrypoint to generate configuration
ENTRYPOINT ["/start-metastore.sh"]