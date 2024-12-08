#!/bin/bash

# Create hive-site.xml dynamically
cat > /opt/hive/conf/hive-site.xml << EOF
<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type="text/xsl" href="configuration.xsl"?>
<configuration>
    <property>
        <name>javax.jdo.option.ConnectionDriverName</name>
        <value>$METASTORE_DB_DRIVER</value>
    </property>
    <property>
        <name>javax.jdo.option.ConnectionURL</name>
        <value>$METASTORE_DB_URL</value>
    </property>
    <property>
        <name>javax.jdo.option.ConnectionUserName</name>
        <value>$METASTORE_DB_USER</value>
    </property>
    <property>
        <name>javax.jdo.option.ConnectionPassword</name>
        <value>$METASTORE_DB_PASSWORD</value>
    </property>
    <property>
    <name>hive.metastore.schema.verification</name>
    <value>false</value>
    <description>
      Enforce metastore schema version consistency.
      True: Verify that version information stored in is compatible with one from Hive jars.  Also disable automatic
            schema migration attempt. Users are required to manually migrate schema after Hive upgrade which ensures
            proper metastore schema migration. (Default)
      False: Warn if the version information stored in metastore doesn't match with one from in Hive jars.
    </description>
  </property>
</configuration>
EOF

# Print the configuration for debugging
echo "Generated hive-site.xml:"
cat /opt/hive/conf/hive-site.xml
schematool -initSchema -dbType postgres
# Start the Hive metastore service
exec hive --service metastore