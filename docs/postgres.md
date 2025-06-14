Postgres
===

psql -U postgres
```
CREATE USER tm;
ALTER USER tm PASSWORD 'tm_password';
CREATE DATABASE topology_manager;
GRANT ALL PRIVILEGES ON DATABASE topology_manager TO tm;
ALTER DATABASE topology_manager OWNER TO tm;
```
