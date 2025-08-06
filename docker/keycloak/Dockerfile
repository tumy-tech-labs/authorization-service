FROM quay.io/keycloak/keycloak:24.0.4

# Copy realm export file
COPY realm-export.json /opt/keycloak/data/import/realm-export.json

# Start Keycloak in development mode and import the realm
CMD ["start-dev", "--import-realm"]
